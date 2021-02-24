## Development Flow

1. Branch
    - We prefer `<category>/<name>` branch.  
      e.g., `feat/merge-automation`, `fix/configs-bug`  
      (Refer to [the link](https://github.com/pvdlg/conventional-changelog-metahub#commit-types) for the category candidates)
2. Commit
    - Follow [the link](https://chris.beams.io/posts/git-commit/) for the good commit messages
3. Pull Request
    - Add commits until the CI pipelines succeed
    - Get approval from one or more maintainers
4. Merge
    - If the author of the pull request is a member of tmax-cloud, the author should merge it
    - If not, one of the approver should take the responsibility
    - Select `Squash and Merge` for a linear and clean commit tree
