package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

var (
	mainCmd = &cobra.Command{
		Use:   "docker-build",
		Short: "Build a docker image for a PR or Commit",
		Long: `Builds a docker image for a PR or Commit.
For commits, the docker image will be pushed if the build is successful`,
		RunE: handleMainCmd,
	}
	prCmd = &cobra.Command{
		Use:   "pr",
		Short: "Build a docker image for a PR",
		Long: `Builds a docker image for a PR.
All layers will be built, but the image will not be pushed`,
		RunE: handlePrCmd,
	}
	commitCmd = &cobra.Command{
		Use:   "commit",
		Short: "Build a docker image for a commit",
		Long: `Builds a docker image for a commit.
Builds all layers and pushes the image to a registry if successful`,
		RunE: handleCommitCmd,
	}
)

func configureCmds() {
	prFlags := prCmd.Flags()

	prFlags.String("repo", "", "the repo clone url")
	prCmd.MarkFlagRequired("repo")

	prFlags.String("clone-path", "", "the path to clone the repo to")
	prCmd.MarkFlagRequired("clone-path")

	prFlags.String("pr-revision-hash", "", "the PR revision id (e.g. commit sha hash)")
	prCmd.MarkFlagRequired("pr-revision-hash")

	prFlags.String("pr-revision-ref", "", "the PR ref that will be used locally")
	prCmd.MarkFlagRequired("pr-revision-ref")

	prFlags.String("base-revision-hash", "", "the base ref revision id (e.g. commit sha hash)")
	prCmd.MarkFlagRequired("base-revision-hash")

	prFlags.String("base-revision-ref", "", "the base ref that will be used locally")
	prCmd.MarkFlagRequired("base-revision-ref")

	prFlags.String("dockerfile", "", "the path to the dockerfile to build")
	prCmd.MarkFlagRequired("dockerfile")

	prFlags.String("docker-context-dir", "", "the path to the docker context used for the build")
	prCmd.MarkFlagRequired("docker-context-dir")

	commitFlags := commitCmd.Flags()

	commitFlags.String("repo", "", "the repo clone url")
	commitCmd.MarkFlagRequired("repo")

	commitFlags.String("clone-path", "", "the path to clone the repo to")
	commitCmd.MarkFlagRequired("clone-path")

	commitFlags.String("revision-hash", "", "the revision id (e.g. commit sha hash)")
	commitCmd.MarkFlagRequired("revision-hash")

	commitFlags.String("revision-ref", "", "the ref that will be used locally")
	commitCmd.MarkFlagRequired("revision-ref")

	commitFlags.String("dockerfile", "", "the path to the dockerfile to build")
	commitCmd.MarkFlagRequired("dockerfile")

	commitFlags.String("docker-context-dir", "", "the path to the docker context used for the build")
	commitCmd.MarkFlagRequired("docker-context-dir")

	mainCmd.AddCommand(prCmd, commitCmd)
}

func handleMainCmd(cmd *cobra.Command, ards []string) error {
	return fmt.Errorf("Must specify a subcommand")
}

func handlePrCmd(cmd *cobra.Command, ards []string) error {
	// Parse command flags
	prFlags := cmd.Flags()

	repoCloneUrl, err := prFlags.GetString("repo")
	if err != nil {
		return fmt.Errorf("error processing pr repo flag")
	}

	clonePath, err := prFlags.GetString("clone-path")
	if err != nil {
		return fmt.Errorf("error processing pr clone-path flag")
	}

	prRevisionHash, err := prFlags.GetString("pr-revision-hash")
	if err != nil {
		return fmt.Errorf("error processing pr-revision-hash flag")
	}

	prRevisionRef, err := prFlags.GetString("pr-revision-ref")
	if err != nil {
		return fmt.Errorf("error processing pr-revision-ref flag")
	}

	baseRevisionHash, err := prFlags.GetString("base-revision-hash")
	if err != nil {
		return fmt.Errorf("error processing base-revision-hash flag")
	}

	baseRevisionRef, err := prFlags.GetString("base-revision-ref")
	if err != nil {
		return fmt.Errorf("error processing base-revision-ref flag")
	}

	dockerfile, err := prFlags.GetString("dockerfile")
	if err != nil {
		return fmt.Errorf("error processing pr dockerfile flag")
	}

	dockerContextDir, err := prFlags.GetString("docker-context-dir")
	if err != nil {
		return fmt.Errorf("error processing pr docker-context-dir flag")
	}

	// Print command flags
	fmt.Printf("PR build with params:\n")
	fmt.Printf("- repo: %s\n", repoCloneUrl)
	fmt.Printf("- clonePath: %s\n", clonePath)
	fmt.Printf("- prRevisionHash: %s\n", prRevisionHash)
	fmt.Printf("- prRevisionRef: %s\n", prRevisionRef)
	fmt.Printf("- baseRevisionHash: %s\n", baseRevisionHash)
	fmt.Printf("- baseRevisionRef: %s\n", baseRevisionRef)
	fmt.Printf("- dockerfile: %s\n", dockerfile)
	fmt.Printf("- dockerContextDir: %s\n", dockerContextDir)

	// Run command
	// Initialize repo
	initOptions := git.PlainInitOptions{
		InitOptions: git.InitOptions{
			DefaultBranch: plumbing.ReferenceName(baseRevisionRef),
		},
	}
	repo, err := git.PlainInitWithOptions(clonePath, &initOptions)
	if err != nil {
		return fmt.Errorf("error in pr git init: %s", err)
	}

	remoteConfig := config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{repoCloneUrl},
	}
	remote, err := repo.CreateRemote(&remoteConfig)
	if err != nil {
		return fmt.Errorf("error in pr git remote config: %s", err)
	}

	// Fetch base and pr commits
	baseRefSpec := fmt.Sprintf("%s:%s", baseRevisionHash, baseRevisionRef)
	baseFetchOptions := git.FetchOptions{
		Depth:    1,
		RefSpecs: []config.RefSpec{config.RefSpec(baseRefSpec)},
		Tags:     git.NoTags,
	}
	err = remote.Fetch(&baseFetchOptions)
	if err != nil {
		return fmt.Errorf("error in pr git fetch base: %s", err)
	}

	prRefSpec := fmt.Sprintf("%s:%s", prRevisionHash, prRevisionRef)
	prFetchOptions := git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec(prRefSpec)},
		Tags:     git.NoTags,
	}
	err = remote.Fetch(&prFetchOptions)
	if err != nil {
		return fmt.Errorf("error in pr git fetch pr: %s", err)
	}

	// Checkout pr commit
	workTree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error in pr git worktree: %s", err)
	}
	checkoutOptions := git.CheckoutOptions{
		Branch: plumbing.ReferenceName(prRevisionRef),
	}
	err = workTree.Checkout(&checkoutOptions)
	if err != nil {
		return fmt.Errorf("error in pr git checkout: %s", err)
	}

	// Check diff between pr and base commits
	baseCommit, err := repo.CommitObject(plumbing.NewHash(baseRevisionHash))
	if err != nil {
		return fmt.Errorf("error in base git commit: %s", err)
	}
	prCommit, err := repo.CommitObject(plumbing.NewHash(prRevisionHash))
	if err != nil {
		return fmt.Errorf("error in pr git commit: %s", err)
	}
	patch, err := baseCommit.Patch(prCommit)
	if err != nil {
		return fmt.Errorf("error in pr git patch: %s", err)
	}

	if shouldBuildImage(patch, dockerfile, dockerContextDir) {
		buildImage(dockerfile, dockerContextDir)
	}

	return nil
}

func handleCommitCmd(cmd *cobra.Command, ards []string) error {
	// Parse command flags
	commitFlags := cmd.Flags()

	repoCloneUrl, err := commitFlags.GetString("repo")
	if err != nil {
		return fmt.Errorf("error processing commit repo flag")
	}

	clonePath, err := commitFlags.GetString("clone-path")
	if err != nil {
		return fmt.Errorf("error processing commit clone-path flag")
	}

	revisionHash, err := commitFlags.GetString("revision-hash")
	if err != nil {
		return fmt.Errorf("error processing commit revision-hash flag")
	}

	revisionRef, err := commitFlags.GetString("revision-ref")
	if err != nil {
		return fmt.Errorf("error processing commit revision-ref flag")
	}

	dockerfile, err := commitFlags.GetString("dockerfile")
	if err != nil {
		return fmt.Errorf("error processing commit dockerfile flag")
	}

	dockerContextDir, err := commitFlags.GetString("docker-context-dir")
	if err != nil {
		return fmt.Errorf("error processing commit docker-context-dir flag")
	}

	// Print command flags
	fmt.Printf("Commmit build with params:\n")
	fmt.Printf("- repo: %s\n", repoCloneUrl)
	fmt.Printf("- clonePath: %s\n", clonePath)
	fmt.Printf("- revisionHash: %s\n", revisionHash)
	fmt.Printf("- revisionRef: %s\n", revisionRef)
	fmt.Printf("- dockerfile: %s\n", dockerfile)
	fmt.Printf("- dockerContextDir: %s\n", dockerContextDir)

	// Run command
	// Initialize repo
	initOptions := git.PlainInitOptions{
		InitOptions: git.InitOptions{
			DefaultBranch: plumbing.ReferenceName(revisionRef),
		},
	}
	repo, err := git.PlainInitWithOptions(clonePath, &initOptions)
	if err != nil {
		return fmt.Errorf("error in commit git init: %s", err)
	}

	remoteConfig := config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{repoCloneUrl},
	}
	remote, err := repo.CreateRemote(&remoteConfig)
	if err != nil {
		return fmt.Errorf("error in commit git remote config: %s", err)
	}

	// Fetch commit
	refSpec := fmt.Sprintf("%s:%s", revisionHash, revisionRef)
	fetchOptions := git.FetchOptions{
		Depth:    2,
		RefSpecs: []config.RefSpec{config.RefSpec(refSpec)},
		Tags:     git.NoTags,
	}
	err = remote.Fetch(&fetchOptions)
	if err != nil {
		return fmt.Errorf("error in commit git fetch pr: %s", err)
	}

	// Checkout commit
	workTree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error in commit git worktree: %s", err)
	}
	checkoutOptions := git.CheckoutOptions{
		Branch: plumbing.ReferenceName(revisionRef),
	}
	err = workTree.Checkout(&checkoutOptions)
	if err != nil {
		return fmt.Errorf("error in commit git checkout: %s", err)
	}

	// Check commit diff
	commit, err := repo.CommitObject(plumbing.NewHash(revisionHash))
	if err != nil {
		return fmt.Errorf("error in git commit: %s", err)
	}
	if commit.NumParents() != 1 {
		return fmt.Errorf(
			"expected commit (%s) to have one parent but found: %d",
			revisionHash,
			commit.NumParents(),
		)
	}

	parent, err := commit.Parents().Next()
	if err != nil {
		return fmt.Errorf("error in git commit parent: %s", err)
	}

	patch, err := parent.Patch(commit)
	if err != nil {
		return fmt.Errorf("error in git commit patch: %s", err)
	}

	if shouldBuildImage(patch, dockerfile, dockerContextDir) {
		buildImage(dockerfile, dockerContextDir)
	}

	return nil
}

func shouldBuildImage(patch *object.Patch, dockerfile string, dockerContextDir string) bool {
	dockerfile = strings.TrimSpace(dockerfile)
	dockerContextDir = strings.TrimSpace(dockerContextDir)

	if dockerContextDir == "" {
		fmt.Println("Building image since dockerContextDir is empty")
		return true
	}

	changedPaths := make(map[string]struct{})
	for _, filePatch := range patch.FilePatches() {
		from, to := filePatch.Files()
		if from != nil {
			changedPaths[from.Path()] = struct{}{}
		}
		if to != nil {
			changedPaths[to.Path()] = struct{}{}
		}
	}

	fmt.Printf("Changed paths:\n")
	for changedPath := range changedPaths {
		fmt.Printf("- %s\n", changedPath)
	}

	for changedPath := range changedPaths {
		if strings.HasPrefix(changedPath, dockerContextDir) {
			return true
		}
		if changedPath == dockerfile {
			return true
		}
	}

	fmt.Println("Skip image build due to no matching files")
	return false
}

func buildImage(dockerfile string, dockerContextDir string) {
}

func main() {
	configureCmds()
	if err := mainCmd.Execute(); err != nil {
		fmt.Printf("error executing command: %s\n", err)
		os.Exit(1)
	}
}
