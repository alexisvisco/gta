package diff

import (
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/gitignore"
	"gopkg.in/src-d/go-git.v4/utils/merkletrie"
	"gopkg.in/src-d/go-git.v4/utils/merkletrie/filesystem"
	"gopkg.in/src-d/go-git.v4/utils/merkletrie/noder"

	"github.com/alexisvisco/gta/pkg/gta/packages"
	"github.com/alexisvisco/gta/pkg/gta/vars"
)

const localRef = ""

func Diff(repository *git.Repository, previousRef, currentRef string) (packages.Packages, error) {
	previousNoder, err := getTree(repository, previousRef)
	if err != nil {
		return nil, err
	}

	if currentRef == localRef {
		return localDiff(vars.Repository, previousNoder)
	} else {
		currentNoder, err := getTree(repository, currentRef)
		if err != nil {
			return nil, err
		}

		return diff(vars.Repository, previousNoder, currentNoder)
	}
}

func localDiff(repo *git.Repository, previous noder.Noder) (packages.Packages, error) {
	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	submodules, err := getSubmodulesStatus(wt)
	if err != nil {
		return nil, err
	}
	current := filesystem.NewRootNode(wt.Filesystem, submodules)

	return diff(repo, previous, current)
}

func getSubmodulesStatus(w *git.Worktree) (map[string]plumbing.Hash, error) {
	o := map[string]plumbing.Hash{}

	sub, err := w.Submodules()
	if err != nil {
		return nil, err
	}

	status, err := sub.Status()
	if err != nil {
		return nil, err
	}

	for _, s := range status {
		if s.Current.IsZero() {
			o[s.Path] = s.Expected
			continue
		}

		o[s.Path] = s.Current
	}

	return o, nil
}

func diff(repo *git.Repository, previous noder.Noder, current noder.Noder) (packages.Packages, error) {
	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	changes, err := merkletrie.DiffTree(previous, current, diffTreeIsEquals)
	if err != nil {
		return nil, err
	}

	return packages.FromChanges(excludeIgnoredChanges(wt, changes)), nil
}

func excludeIgnoredChanges(w *git.Worktree, changes merkletrie.Changes) merkletrie.Changes {
	patterns, err := gitignore.ReadPatterns(w.Filesystem, nil)
	if err != nil {
		return changes
	}

	patterns = append(patterns, w.Excludes...)

	if len(patterns) == 0 {
		return changes
	}

	m := gitignore.NewMatcher(patterns)

	var res merkletrie.Changes
	for _, ch := range changes {
		var path []string
		for _, n := range ch.To {
			path = append(path, n.Name())
		}
		if len(path) == 0 {
			for _, n := range ch.From {
				path = append(path, n.Name())
			}
		}
		if len(path) != 0 {
			isDir := (len(ch.To) > 0 && ch.To.IsDir()) || (len(ch.From) > 0 && ch.From.IsDir())
			if m.Match(path, isDir) {
				continue
			}
		}
		res = append(res, ch)
	}
	return res
}
