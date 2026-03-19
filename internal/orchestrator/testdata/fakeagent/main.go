// fakeagent is a test helper that simulates a planning agent for end-to-end
// pipeline tests. It reads .doug/plan/ACTIVE_STEP.md to determine the current
// stage, writes the appropriate planning artifact, and records SUCCESS.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// epicIDsFromRoadmap extracts epic IDs (e.g. "EPIC-1") from a ROADMAP.md
// by scanning for lines of the form "id: EPIC-N".
func epicIDsFromRoadmap(roadmap string) []string {
	var ids []string
	for _, line := range strings.Split(roadmap, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "id: ") {
			id := strings.TrimPrefix(line, "id: ")
			if strings.HasPrefix(id, "EPIC-") {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fakeagent: getwd:", err)
		os.Exit(1)
	}

	planDir := filepath.Join(root, ".doug", "plan")
	stepPath := filepath.Join(planDir, "ACTIVE_STEP.md")

	data, err := os.ReadFile(stepPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fakeagent: read ACTIVE_STEP.md:", err)
		os.Exit(1)
	}
	content := string(data)

	outcome := "SUCCESS"

	switch {
	case strings.Contains(content, "**Stage**: Discovery"):
		if err := os.WriteFile(filepath.Join(planDir, "VISION.md"), []byte(fakeVision), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "fakeagent: write VISION.md:", err)
			os.Exit(1)
		}
	case strings.Contains(content, "**Stage**: Roadmapping"):
		if err := os.WriteFile(filepath.Join(planDir, "ROADMAP.md"), []byte(fakeRoadmap), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "fakeagent: write ROADMAP.md:", err)
			os.Exit(1)
		}
	case strings.Contains(content, "**Stage**: Scoping"):
		roadmapData, err := os.ReadFile(filepath.Join(planDir, "ROADMAP.md"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "fakeagent: read ROADMAP.md:", err)
			os.Exit(1)
		}
		epicIDs := epicIDsFromRoadmap(string(roadmapData))
		if len(epicIDs) == 0 {
			fmt.Fprintln(os.Stderr, "fakeagent: no epic IDs found in ROADMAP.md")
			os.Exit(1)
		}
		// Find the next epic that has not yet been scoped.
		var nextEpic string
		for _, id := range epicIDs {
			if _, err := os.Stat(filepath.Join(planDir, "epics", id, "SCOPED.md")); os.IsNotExist(err) {
				nextEpic = id
				break
			}
		}
		if nextEpic == "" {
			// All epics are already scoped — write the completion marker.
			if err := os.WriteFile(filepath.Join(planDir, "SCOPED.md"), []byte(fakeScopedComplete), 0o644); err != nil {
				fmt.Fprintln(os.Stderr, "fakeagent: write SCOPED.md:", err)
				os.Exit(1)
			}
		} else {
			epicDir := filepath.Join(planDir, "epics", nextEpic)
			if err := os.MkdirAll(epicDir, 0o755); err != nil {
				fmt.Fprintln(os.Stderr, "fakeagent: mkdir epics dir:", err)
				os.Exit(1)
			}
			scopedContent := fmt.Sprintf("# Scoped: %s\n", nextEpic)
			if err := os.WriteFile(filepath.Join(epicDir, "SCOPED.md"), []byte(scopedContent), 0o644); err != nil {
				fmt.Fprintln(os.Stderr, "fakeagent: write per-epic SCOPED.md:", err)
				os.Exit(1)
			}
			// Check whether all epics are now scoped.
			allScoped := true
			for _, id := range epicIDs {
				if _, err := os.Stat(filepath.Join(planDir, "epics", id, "SCOPED.md")); os.IsNotExist(err) {
					allScoped = false
					break
				}
			}
			if allScoped {
				if err := os.WriteFile(filepath.Join(planDir, "SCOPED.md"), []byte(fakeScopedComplete), 0o644); err != nil {
					fmt.Fprintln(os.Stderr, "fakeagent: write SCOPED.md:", err)
					os.Exit(1)
				}
			} else {
				outcome = "RETRY"
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "fakeagent: unrecognised stage in ACTIVE_STEP.md\n%s\n", content)
		os.Exit(1)
	}

	updated := strings.ReplaceAll(content, `outcome: ""`, `outcome: "`+outcome+`"`)
	if err := os.WriteFile(stepPath, []byte(updated), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "fakeagent: update ACTIVE_STEP.md:", err)
		os.Exit(1)
	}
}

const fakeScopedComplete = `# Scoping Complete

All epics from ROADMAP.md have been scoped. Scoped definitions are in ` + "`.doug/plan/epics/`." + `
`

const fakeVision = `# Vision

## Project Name

Habify

## Problem Statement

People struggle to build and maintain habits because they lack a simple,
non-judgmental tool that tracks daily progress without overwhelming them
with features. Most habit trackers become another source of anxiety rather
than a source of motivation.

## Target Users

Adults aged 20-45 who want to establish and maintain personal habits,
particularly those who have tried and abandoned other habit-tracking apps
due to complexity or guilt-inducing streaks.

## Goals

- Allow users to create and track up to ten daily habits with a single-tap check-in.
- Provide a 30-day streak calendar to reinforce consistency.
- Send one optional daily reminder notification per habit.

## Non-Goals

- Social or sharing features in v1.
- Habit templates or a community library.
- Integration with wearables or third-party health apps.
- Web browser version.

## Constraints

- iOS and Android only, built with React Native for cross-platform efficiency.
- Launch within three months with a two-person development team.
- Free tier only in v1; no backend required (local storage).

## Success Criteria

- 500 daily active users within 60 days of launch.
- Average session length under two minutes.
- 4.0 or higher star rating on both app stores within 30 days.

## Failure Conditions

- App crashes or data loss on any tested device.
- Onboarding requires more than three steps to create the first habit.
- Users report feeling judged or anxious when they miss a day.

## Background

Research shows that habit formation requires consistent tracking and positive
reinforcement without shame. Existing apps such as Habitica and Streaks either
over-gamify the experience or feel clinical. Habify targets the middle ground:
minimal, warm, and forgiving.
`

const fakeRoadmap = `---
project: "Habify"
generated: "2026-03-18"
source: VISION.md
---

# Roadmap

## EPIC-1: Core Habit Tracking

---
id: EPIC-1
name: "Core Habit Tracking"
sequence: 1
status: planned
---

Build the foundational habit-creation and daily check-in flow. This epic
delivers the core user value — a single-tap interaction that persists data
locally on device — and establishes the data model and UI skeleton that all
subsequent epics depend on.

## EPIC-2: Streak Visualization

---
id: EPIC-2
name: "Streak Visualization"
sequence: 2
status: planned
---

Add the 30-day streak calendar and progress visualization. This epic depends
on the data model from EPIC-1 and provides the primary motivation loop: users
see their consistency over time and are encouraged to maintain it.

## EPIC-3: Reminder Notifications

---
id: EPIC-3
name: "Reminder Notifications"
sequence: 3
status: planned
---

Implement optional daily reminder notifications, one per habit. Built after the
core check-in flow is stable, this epic adds the gentle nudge that helps users
remember to log without requiring constant active recall.

## EPIC-4: App Store Launch

---
id: EPIC-4
name: "App Store Launch"
sequence: 4
status: planned
---

Finalize onboarding, complete device testing, and submit to the iOS App Store
and Google Play. This epic gates on all previous epics being complete and
delivers the product to its first users.
`
