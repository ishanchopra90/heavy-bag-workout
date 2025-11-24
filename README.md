# Puppy Power - Heavy Bag Workout App

A Go-based application that generates randomized boxing workouts with combinations of punches and defensive moves. The app provides an interactive CLI interface for heavy bag training with customizable workout patterns, timing, and audio cues.

## Features

- **6 Boxing Punches**: Jab, Cross, Lead Hook, Rear Hook, Lead Uppercut, Rear Uppercut
- **6 Defensive Moves**: Left Slip, Right Slip, Left Roll, Right Roll, Pull Back, Duck
- **Randomized Workouts**: Generates unique combinations of punches and defensive moves
- **Workout Patterns**: Choose from linear, pyramid, random, or constant complexity patterns
- **Stance Support**: Full support for both right-handed (orthodox) and left-handed (southpaw) boxers with stance-specific punch naming
- **Configurable Timing**: Customize work and rest periods per round
- **Preset Configurations**: Quick-start with pre-configured workouts (beta_style, endurance, power)
- **LLM Integration**: Optional AI-powered workout generation using OpenAI's GPT models
- **Audio Cues**: 
  - Beep at configurable intervals during work periods (default: 5 seconds, adjustable via `--tempo`)
  - Voice announcements for period transitions ("work", "rest")
  - Combo callouts at the start of each round
  - "Workout complete" announcement at the end
  - 3 beeps in the last 3 seconds of rest periods to signal readiness
- **Interactive CLI**: Real-time workout display with progress tracking
- **Configuration Files**: JSON-based configuration for custom workout setups

## Project Structure

```
HeavyBagWorkout/
├── cmd/
│   └── heavybagworkout/    # Main application entry point
├── internal/
│   ├── models/             # Data models (Punch, Combo, Workout, etc.)
│   ├── generator/          # Combo and workout generation logic
│   ├── timer/              # Timer functionality and audio cues
│   ├── config/             # Configuration management
│   ├── cli/                # CLI interface
│   └── mocks/              # Generated mocks for testing
├── configs/                # Preset configuration files
├── go.mod
├── go.sum
└── README.md
```

## Setup Instructions

### Prerequisites

- Go 1.19 or higher
- (Optional) OpenAI API key for LLM-powered workout generation

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd HeavyBagWorkout
```

2. Build the application:
```bash
go build -o heavybagworkout ./cmd/heavybagworkout
```

3. Run the application:
```bash
./heavybagworkout
```

Or run directly with Go:
```bash
go run ./cmd/heavybagworkout
```

## Usage

### Quick Start

Start with a preset configuration:
```bash
./heavybagworkout --preset beta_style
```

### Command-Line Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--config` | Path to JSON configuration file | `--config configs/custom.json` |
| `--preset` | Use a preset (beta_style, endurance, power) | `--preset power` |
| `--work-duration` | Work period duration in seconds | `--work-duration 30` |
| `--rest-duration` | Rest period duration in seconds | `--rest-duration 15` |
| `--rounds` | Total number of rounds | `--rounds 10` |
| `--pattern` | Workout pattern (linear, pyramid, random, constant) | `--pattern pyramid` |
| `--min-moves` | Minimum moves per combo | `--min-moves 2` |
| `--max-moves` | Maximum moves per combo | `--max-moves 6` |
| `--include-defensive` | Include defensive moves in combos | `--include-defensive` |
| `--no-include-defensive` | Disable defensive moves | `--no-include-defensive` |
| `--use-llm` | Use LLM for workout generation | `--use-llm` |
| `--openai-api-key` | OpenAI API key | `--openai-api-key sk-...` |
| `--stance` | Boxer's stance (orthodox, southpaw) | `--stance southpaw` |
| `--tempo` | Workout tempo: Slow (5s), Medium (4s), Fast (3s), Superfast (2s) | `--tempo fast` |
| `--version` | Show version information | `--version` |
| `--help` | Show help message | `--help` |

### Configuration Priority

1. **Command-line flags** (highest priority) - Override all other settings
2. **Config file** (`--config`) - Load from custom JSON file
3. **Preset** (`--preset`) - Use predefined configuration
4. **Default configuration** (lowest priority) - Built-in defaults

### Examples

**Basic workout with preset:**
```bash
./heavybagworkout --preset beta_style
```

**Custom workout with specific timing:**
```bash
./heavybagworkout --work-duration 30 --rest-duration 15 --rounds 10
```

**Pyramid pattern with defensive moves:**
```bash
./heavybagworkout --pattern pyramid --min-moves 2 --max-moves 6 --include-defensive
```

**Southpaw stance with LLM generation:**
```bash
./heavybagworkout --preset power --stance southpaw --use-llm
```

**Using a custom configuration file:**
```bash
./heavybagworkout --config configs/custom.json
```

## Workout Patterns

The app supports four workout patterns that control how combo complexity varies across rounds:

- **Linear**: Combo complexity increases linearly from round 1 to the final round
  - Round 1: Simpler combos (closer to min moves)
  - Final round: More complex combos (closer to max moves)

- **Pyramid**: Combo complexity peaks in the middle rounds
  - Starts simple, increases to peak in middle rounds, then decreases
  - Great for building intensity and then tapering

- **Random**: Combo complexity varies randomly within the min-max range
  - Unpredictable and keeps you on your toes
  - Each round can be different

- **Constant**: Combo complexity remains relatively constant throughout
  - Consistent challenge level across all rounds
  - Good for endurance training

## Stance Support

The app fully supports both boxing stances:

- **Orthodox** (right-handed): Default stance
  - Jab = left hand, Cross = right hand
  - Lead Hook = left hook, Rear Hook = right hook
  - Lead Uppercut = left uppercut, Rear Uppercut = right uppercut

- **Southpaw** (left-handed): Mirror stance
  - Jab = right hand, Cross = left hand
  - Lead Hook = right hook, Rear Hook = left hook
  - Lead Uppercut = right uppercut, Rear Uppercut = left uppercut

Defensive moves are automatically paired appropriately with punches based on the selected stance.

## LLM Integration

The app can use OpenAI's GPT models to generate more creative and varied workouts:

1. Set your OpenAI API key:
   ```bash
   export OPENAI_API_KEY=sk-your-key-here
   ```
   Or use the `--openai-api-key` flag

2. Enable LLM generation:
   ```bash
   ./heavybagworkout --use-llm --preset power
   ```

The LLM generator understands:
- Stance-specific punch naming
- Defensive move pairing with punches
- Workout patterns and complexity requirements
- Realistic boxing combinations

## Audio Cues

The app provides audio feedback during workouts:

- **Beep at configurable intervals** during work periods as a reminder to execute the combo
  - Default: 5 seconds (Slow tempo)
  - Adjustable via `--tempo` flag: Slow (5s), Medium (4s), Fast (3s), Superfast (2s)
- **"Work"** voice announcement when transitioning to a work period
- **"Rest"** voice announcement when transitioning to a rest period
- **3 beeps** in the last 3 seconds of rest periods to signal readiness for the next work period
- **Combo callout** at the start of each round (speaks the moves)
- **"Workout complete"** announcement when the workout finishes

Audio cues use system text-to-speech and are enabled by default.

## Configuration Files

Configuration files are JSON format. Example structure:

```json
{
  "workout": {
    "work_duration_seconds": 30,
    "rest_duration_seconds": 15,
    "total_rounds": 8
  },
  "pattern": {
    "type": "pyramid",
    "min_moves": 2,
    "max_moves": 7,
    "include_defensive": true
  },
  "generator": {
    "use_llm": false,
    "llm_model": "gpt-4.1-nano"
  },
  "stance": "orthodox",
  "openai_api_key": ""
}
```

### Available Presets

- **beta_style**: Quick, high-intensity rounds
- **endurance**: Longer rounds for stamina building
- **power**: Balanced rounds with defensive moves

## Interactive Controls

During a workout:

- **Ctrl+C**: Cancel the workout at any time
- **Enter**: View workout preview before starting

Note: Pause/resume functionality will be available in the future GUI version.

## Development Status

- **V1** - CLI (MVP is out)
- **V2** - GUI with animations of Scrappy Doo in fighting poses(in development...)

## License

This project is licensed under the MIT License.

Copyright (c) 2025 Ishan Chopra

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
