# Test Data for Za

This directory contains realistic test data for the Za (Zettelkasten Augmentation) tool.

## Structure

```
testdata/
├── journal/          # Daily journal entries
├── standup/          # Daily standup notes
└── .za.yaml         # Sample configuration file
```

## Test Data Coverage

The test data spans from **November 25, 2024 to February 5, 2025** (~2.5 months) and includes:

### Date Coverage

**November 2024** (Pre-holiday):
- **2024-11-25** (Monday) - Full journal + standup
- **2024-11-26 to 2024-12-18** - **GAP** (22 days - tests 30-day window boundary)

**December 2024** (Holiday period):
- **2024-12-19** (Thursday) - Full journal + standup (last day before holiday)
- **2024-12-20** (Friday) - Full journal + standup (holiday starts)
- **2024-12-21 to 2025-01-05** - **GAP** (16 days - holiday break)

**January 2025** (Main work period):
- **2025-01-06** (Monday) - Full journal + standup (return from holiday)
- **2025-01-07** (Tuesday) - Full journal + standup
- **2025-01-08** (Wednesday) - Full journal + standup
- **2025-01-09** (Thursday) - **GAP** (1 day - tests single-day fallback)
- **2025-01-10** (Friday) - Full journal + standup
- **2025-01-11/12** (Weekend) - **GAP** (tests weekend handling)
- **2025-01-13** (Monday) - Full journal + standup
- **2025-01-14** (Tuesday) - Full journal + standup
- **2025-01-15 to 2025-02-02** - **GAP** (19 days - holiday to Australia)

**February 2025** (Return from holiday & Goals copying):
- **2025-02-03** (Monday) - Full journal + standup (return from holiday, Week 1 starts)
- **2025-02-04** (Tuesday) - Full journal + standup (Week 1)
- **2025-02-05** (Wednesday) - Full journal + standup (Week 1)
- **2025-02-06 to 2025-02-09** - **GAP** (4 days - Thu-Sun, tests week boundary)
- **2025-02-10** (Monday) - Full journal + standup (Week 2 starts)

### Features Demonstrated

#### Markdown Elements
- YAML frontmatter with various tags (including automatic company tags on weekdays)
- Multiple heading levels
- Nested bullet lists (up to 3 levels)
- Code blocks (Go, YAML examples)
- Blockquotes
- Task lists with checkboxes `[ ]` and `[x]`
- Bold, italic formatting

#### Link Types
1. **Wiki-style links**: `[[person-name]]`, `[[project-reference]]`
2. **Relative date links**: `[Yesterday](2025-01-06)`, `[Tomorrow](2025-01-08)`
3. **Cross-reference links**: `[Standup](../standup/2025-01-06)`, `[Daily](../journal/2025-01-06)`
4. **External links**:
   - Linear issues: `https://linear.app/acme/issue/PLA-482`
   - GitHub: `https://github.com/acme-co/auth-service/pull/142`
   - Notion: `https://www.notion.so/acme/...`
   - Google Meet: `https://meet.google.com/...`
   - RFCs: `https://datatracker.ietf.org/doc/html/rfc7662`

#### Sections in Journal Entries
- **Goals of the Week** - Weekly objectives (copied within same week)
- **Goals of the Day** - Daily tasks with checkboxes and plain bullets
- Work Completed
- Worked On
- Meetings (with attendees and notes)
- Calendar (with meeting details)
- Thoughts
- Links

#### Sections in Standup Entries
- Worked on Yesterday
- Working on Today
- Blocked on
- Notes
- Links

### Edge Cases Tested

1. **Single-Day Gaps**: Missing Thursday (2025-01-09) tests backward search for adjacent days
2. **Weekend Gaps**: Missing Saturday/Sunday tests weekend handling
3. **Holiday Break**: 16-day gap (Dec 21 - Jan 5) tests multi-week holiday scenarios
4. **Extended Holiday**: 19-day gap (Jan 15 - Feb 2) tests longer absence periods
5. **30-Day Window Boundary**: Nov 25 entry is ~22 days before Dec 19, testing near-boundary search
6. **Stale Links**: Links reference "Yesterday"/"Tomorrow" that may need updating across gaps
7. **Inconsistent Sections**: Not all entries have all sections
8. **Various Link Formats**: Multiple ways of referencing dates and notes
9. **Mixed Content**: Technical content (code, commands) mixed with narrative
10. **External References**: Various external systems referenced
11. **Return-to-Work Context**: Entries after long gaps reference catching up and reviewing missed work
12. **Week Boundary**: Feb 10 (Monday) tests new week vs. continuing same week
13. **Goals Copying**: Feb 3-5, 10 demonstrate goals copying behavior
14. **Company Tags**: Weekday entries (Mon-Fri) include `company:acme` tag automatically

## Test Scenarios

### Extraction Tests
- Extract "Work Completed" from journal entries
- Extract "Worked On" from journal entries
- Extract "Worked on Yesterday" from standup entries
- Handle missing sections gracefully
- Preserve formatting in extracted content

### Link Detection Tests
- Detect relative date links that need updating
- Find cross-references between journal and standup
- Identify wiki-style links to people and projects
- Preserve external links unchanged

### Date Search Tests
- Find entry for exact date
- Fall back to previous entry when date missing (e.g., 2025-01-09 → 2025-01-08)
- Search backwards within 30-day window
- Handle weekend gaps appropriately
- Handle multi-week holiday gaps (16-19 days)
- Test 30-day window boundary behavior (entries near the limit)
- Gracefully fail when no entry found within 30-day window

### Link Fixing Tests
- Update "Yesterday" links to actual previous date
- Update "Tomorrow" links to actual next date
- Handle single-day gaps (e.g., Jan 8's "Tomorrow" should point to Jan 10, skipping Jan 9)
- Handle multi-week gaps (e.g., Dec 20's "Return from Holiday" link to Jan 6)
- Update cross-references when dates change
- Handle special link titles like "Return from Holiday", "Last Week", "Before Vacation"

### Goals Copying Tests (Feb 3-5, 10)

#### Goals of the Week
- **Same Week Behavior** (Feb 3-5):
  - Feb 3 (Monday): Establishes Goals of the Week for Week 1
  - Feb 4 (Tuesday): Should copy Goals of the Week from Feb 3 (same week)
  - Feb 5 (Wednesday): Should copy Goals of the Week from Feb 4 (same week)
- **New Week Behavior** (Feb 10):
  - Feb 10 (Monday): Week 2 starts - should NOT copy Goals of the Week from Feb 5
  - Should have new Goals of the Week for the new week

#### Goals of the Day
- **Uncompleted Tasks** `[ ]`: Should be copied forward
- **Completed Tasks** `[x]`: Should NOT be copied forward
- **Plain Bullets** (no checkbox): Should be copied forward (state unknown)
- **Test Progression**:
  - Feb 3: Mixed items - `[ ]` unchecked, `[x]` completed, plain bullets
  - Feb 4: Should copy from Feb 3:
    - ✓ `[ ]` Review 3 weeks of changes
    - ✗ NOT `[x]` Sync with team (completed)
    - ✓ `[ ]` Clear inbox
    - ✓ Plain bullet: Check Slack messages
  - Feb 5: Should copy unfinished from Feb 4
  - Feb 10: Should copy unfinished from Feb 5 (even though it's a new week)

### Company Tag Tests
- **Weekday Entries**: All journal and standup entries on Mon-Fri should have `company:acme` tag
- **Weekend Entries**: Weekend entries (Sat-Sun) should NOT have company tag
- **Configuration**: Company tag is configurable via `company_tag` setting (default: "acme")
- **Format**: Tag is automatically added as `company:{company_tag}` to frontmatter tags array

## Fictional Context

The test data represents a fictional developer (Alex Rivera) working at a fictional company (Acme) on an authentication service project over ~2.5 months. Key aspects:

- **Project**: OAuth2 authentication refactoring with rate limiting and audit logging
- **Team Members**: Alice Smith, Bob Jones, Carol White, Dave Miller, Eve Williams (joins in February)
- **Linear Issues**: PLA-482, PLA-521, PLA-556, PLA-589
- **Technologies**: Go, OAuth2, Redis, PostgreSQL, S3
- **Time Period**: Q4 2024 planning → Holiday break → Q1 2025 implementation → 3-week holiday → Return

### Timeline
1. **Nov 25**: Early planning and research
2. **Dec 19-20**: Pre-holiday wrap-up
3. **Dec 21 - Jan 5**: Holiday break (2+ weeks)
4. **Jan 6-14**: Sprint work - authentication implementation and deployment
5. **Jan 15 - Feb 2**: Holiday to Australia (3 weeks)
6. **Feb 3-5**: Return from holiday, catch up, start audit logging

This provides a realistic scenario with:
- Sprint planning and completion
- Code reviews and pairing sessions
- Production deployment
- Meeting notes and calendar events
- Technical implementation details
- Team collaboration and communication
- **Extended absences** (holidays) with return-to-work context
- Team functioning during absences
- New team member onboarding
