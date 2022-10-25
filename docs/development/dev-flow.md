## Development Flow

1. Branch
    - We prefer `<category>/<name>` branch.  
      e.g., `feat/merge-automation`, `fix/configs-bug`  
      (Refer to [the link](https://github.com/pvdlg/conventional-changelog-metahub#commit-types) for the category candidates)
2. Fix files
    - Try to follow the golang's standard code format
    - Write unit tests if you are implementing a new feature
    - If you are creating a new file, make sure you insert a proper copyright
    - Make sure the tests pass
3. Commit
    - Follow [the link](https://chris.beams.io/posts/git-commit/) for the good commit messages
4. Pull Request
    - Assign yourself for the pull request
    - Request approval from one or more maintainers
    - We prefer a single commit per pull request
    - Update the commit (via force push) until the CI pipelines succeed
6. Merge
    - Your pull request will be merged automatically if it is approved and passes all the required tests
