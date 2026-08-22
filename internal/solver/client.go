package solver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

const (
	defaultBinaryPath   = "solver/target/debug/cryptic"
	defaultWordlistPath = "wordlists/english.txt"
)

var ErrUnavailable = errors.New("cryptic solver unavailable")

type Enumeration struct {
	Raw   string `json:"raw"`
	Parts []int  `json:"parts"`
	Total int    `json:"total"`
}

type Candidate struct {
	Answer         string `json:"answer"`
	Mechanism      string `json:"mechanism"`
	Fodder         string `json:"fodder,omitempty"`
	Pattern        string `json:"pattern,omitempty"`
	Indicator      string `json:"indicator,omitempty"`
	MatchesPattern *bool  `json:"matches_pattern,omitempty"`
}

type Analysis struct {
	Clue        string      `json:"clue"`
	Enumeration Enumeration `json:"enumeration"`
	Candidates  []Candidate `json:"candidates"`
}

type Client struct {
	BinaryPath   string
	WordlistPath string
}

func NewFromEnv() *Client {
	binary := os.Getenv("CRYPTIC_SOLVER_PATH")
	if binary == "" {
		binary = defaultBinaryPath
	}
	wordlist := os.Getenv("CRYPTIC_WORDLIST_PATH")
	if wordlist == "" {
		wordlist = defaultWordlistPath
	}
	return &Client{BinaryPath: binary, WordlistPath: wordlist}
}

func (c *Client) Analyse(ctx context.Context, clue string, known string) (Analysis, error) {
	args := []string{"--wordlist", c.WordlistPath, "analyse", "--clue", clue}
	if known != "" {
		args = append(args, "--known", known)
	}

	output, err := exec.CommandContext(ctx, c.BinaryPath, args...).Output()
	if err != nil {
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			return Analysis{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return Analysis{}, fmt.Errorf("cryptic solver rejected clue: %s", exitErr.Stderr)
		}
		return Analysis{}, fmt.Errorf("run cryptic solver: %w", err)
	}

	var analysis Analysis
	if err := json.Unmarshal(output, &analysis); err != nil {
		return Analysis{}, fmt.Errorf("decode cryptic solver response: %w", err)
	}
	return analysis, nil
}
