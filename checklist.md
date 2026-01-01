# Heavy Bag Workout App - GUI Development Checklist

## Project Setup
1. [x] Add Gio UI dependency to `go.mod`
2. [x] Create GUI application entry point (e.g., `cmd/heavybagworkout-gui/main.go`)
3. [x] Set up Gio UI window and application structure
4. [x] Create GUI package structure (`internal/gui/`)

## Workout Parameter Form
5. [x] Design form layout for workout parameters
6. [x] Add input field for work duration (seconds)
7. [x] Add input field for rest duration (seconds)
8. [x] Add input field for total rounds
9. [x] Add dropdown/selector for workout pattern (linear, pyramid, random, constant)
10. [x] Add input field for minimum moves per combo
11. [x] Add input field for maximum moves per combo
12. [x] Add checkbox/toggle for including defensive moves with text label "Include defensive moves?" next to it
13. [x] Add dropdown/selector for stance (orthodox, southpaw)
14. [x] Add dropdown/selector for tempo (Slow, Medium, Fast, Superfast)
15. [x] Add checkbox/toggle for using LLM generation
16. [x] Add input field for OpenAI API key (optional, masked)
17. [x] Add preset selector (beta_style, endurance, power)
18. [x] Add form validation for all input fields
19. [x] Add "Start Workout" button
20. [x] Add "Load from Config File" button/functionality
21. [x] Add "Save Configuration" button/functionality

## Workout Display Screen
22. [x] Create workout display window/layout
23. [x] Display current round number prominently
24. [x] Display workout progress (rounds completed/total) with progress bar
25. [x] Display countdown timer (MM:SS format)
26. [x] Display current period indicator (Work/Rest)
27. [x] Display current combo moves (formatted list)
28. [x] Add pause/resume button
29. [x] Add stop/quit button
30. [x] Handle window close event (prompt to confirm if workout in progress)
31. [x] Add workout completion screen/message

## Animation System
32. [x] Design/create Scrappy Doo character sprite/asset
33. [x] Create idle/ready pose animation
34. [x] Create jab animation (left hand - orthodox)
35. [x] Create jab animation (right hand - southpaw)
36. [x] Create cross animation (right hand - orthodox)
37. [x] Create cross animation (left hand - southpaw)
38. [x] Create lead hook animation (left hook - orthodox)
39. [x] Create lead hook animation (right hook - southpaw)
40. [x] Create rear hook animation (right hook - orthodox)
41. [x] Create rear hook animation (left hook - southpaw)
42. [x] Create lead uppercut animation (left uppercut - orthodox)
43. [x] Create lead uppercut animation (right uppercut - southpaw)
44. [x] Create rear uppercut animation (right uppercut - orthodox)
45. [x] Create rear uppercut animation (left uppercut - southpaw)
46. [x] Create left slip defensive move animation
47. [x] Create right slip defensive move animation
48. [x] Create left roll defensive move animation
49. [x] Create right roll defensive move animation
50. [x] Create pull back defensive move animation
51. [x] Create duck defensive move animation
52. [x] Create heavy bag visual element/asset
53. [x] Implement animation state machine (idle, punching, defensive, transitions)
54. [x] Sync animations with combo timing (based on tempo setting)
55. [x] Handle animation transitions between moves
56. [x] Add visual feedback for work vs rest periods (different background/colors)

## Backend Integration
57. [x] Integrate GUI with existing workout generator
58. [x] Integrate GUI with existing timer system
59. [x] Integrate GUI with existing audio cue system
60. [x] Pass workout parameters from form to generator
61. [x] Handle workout generation (both LLM and non-LLM paths)
62. [x] Update display based on timer callbacks
63. [x] Handle pause/resume functionality with timer
64. [x] Handle workout completion events

## Audio Integration
65. [x] Ensure beep sounds play during work periods (at tempo intervals)
66. [x] Ensure voice announcements work ("work", "rest", "workout complete")
67. [x] Ensure combo callouts work at round start
68. [x] Ensure 3 beeps play in last 3 seconds of rest periods
69. [ ] Add mute/unmute toggle button in GUI

## Pause/Resume Functionality
70. [ ] Implement pause button click handler
71. [ ] Pause workout timer when pause button clicked
72. [ ] Update UI to show paused state
73. [ ] Disable animations during pause
74. [ ] Implement resume button click handler
75. [ ] Resume workout timer when resume button clicked
76. [ ] Update UI to show active state
77. [ ] Re-enable animations on resume

## Styling and UX
78. [x] Design color scheme/theme
79. [x] Style form inputs consistently
80. [x] Style buttons with hover/active states
81. [ ] Add loading indicators during workout generation
82. [x] Add error messages for invalid inputs
83. [x] Add success/confirmation messages
84. [x] Ensure responsive layout (handle window resizing)
85. [ ] Add keyboard shortcuts (e.g., Space for pause/resume, Q for quit)
86. [x] Add tooltips/help text for form fields

## Testing
87. [x] Test form validation with various inputs
88. [x] Test workout generation with all patterns
89. [x] Test animations for all punch types (orthodox and southpaw)
90. [x] Test animations for all defensive moves
91. [x] Test pause/resume functionality
92. [x] Test audio cues during workout
93. [x] Test workout completion flow
94. [x] Test error handling (invalid API key, network errors, etc.)
95. [x] Test window resize behavior
96. [x] Test on different screen sizes/resolutions

## Asset Management
97. [x] Organize animation assets in `assets/` or `internal/gui/assets/` directory
98. [x] Create asset loading system
99. [x] Optimize animation assets for performance
100. [x] Handle asset loading errors gracefully

## Documentation
101. [x] Update README.md with GUI usage instructions
102. [x] Document GUI-specific features
103. [ ] Add screenshots of GUI to README
104. [x] Document animation system architecture
105. [x] Document asset requirements and formats

## Polish
106. [ ] Add smooth transitions between screens
107. [ ] Add visual effects (e.g., impact effects when punches land)
108. [ ] Optimize animation rendering performance
109. [ ] Add sound effects for punches (optional)
110. [ ] Add visual feedback for combo execution
111. [ ] Test and fix any visual glitches
112. [ ] Ensure consistent frame rate for animations

## Misc Todos/Bugs
113. [ ] Add ability to create workout plans with several saved workouts
114. [x] Add ability to generate audio log for workout plans so it can be played on earphones with smartphone while working out
115. [x] Voice should call out round number - "round 1 of 50 and so on"
