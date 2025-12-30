package gui

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

// CharacterRenderer handles rendering of the Scrappy Doo character
type CharacterRenderer struct {
	character *CharacterSprite
}

// NewCharacterRenderer creates a new character renderer
func NewCharacterRenderer(character *CharacterSprite) *CharacterRenderer {
	return &CharacterRenderer{
		character: character,
	}
}

// Layout renders the character at the current animation frame
func (cr *CharacterRenderer) Layout(gtx layout.Context) layout.Dimensions {
	// Get current animation frame
	frame := cr.character.GetCurrentFrame()
	if frame == nil {
		// No frame available, return empty dimensions
		return layout.Dimensions{}
	}

	// If we have an actual sprite image, render it
	if frame.Image != nil {
		return cr.renderSprite(gtx, frame.Image)
	}

	// Otherwise, render placeholder
	return cr.renderPlaceholder(gtx)
}

// renderSprite renders an actual sprite image
func (cr *CharacterRenderer) renderSprite(gtx layout.Context, img image.Image) layout.Dimensions {
	// Get image bounds
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Calculate scale to fit within available space while maintaining aspect ratio
	maxWidth := gtx.Constraints.Max.X
	maxHeight := gtx.Dp(unit.Dp(300)) // Max character height (increased for better visibility)

	// Calculate scale factors
	scaleX := float32(maxWidth) / float32(imgWidth)
	scaleY := float32(maxHeight) / float32(imgHeight)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Ensure scale doesn't exceed 1.0 (don't upscale)
	if scale > 1.0 {
		scale = 1.0
	}

	scaledWidth := int(float32(imgWidth) * scale)
	scaledHeight := int(float32(imgHeight) * scale)

	// Center the image horizontally
	offsetX := (gtx.Constraints.Max.X - scaledWidth) / 2
	offsetY := gtx.Dp(unit.Dp(20))

	// Apply transforms in correct order: offset first, then scale
	// This ensures the image is positioned correctly, then scaled
	offset := op.Offset(image.Point{X: offsetX, Y: offsetY}).Push(gtx.Ops)
	defer offset.Pop()

	// Apply scale transformation from origin
	scaleTransform := op.Affine(f32.Affine2D{}.Scale(f32.Point{}, f32.Point{X: scale, Y: scale})).Push(gtx.Ops)
	defer scaleTransform.Pop()

	// Clip to the original image bounds (before scaling)
	clipStack := clip.Rect(image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: imgWidth, Y: imgHeight},
	}).Push(gtx.Ops)
	defer clipStack.Pop()

	// Create and paint the image (scaling will be applied by the transform)
	imgOp := paint.NewImageOp(img)
	imgOp.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	return layout.Dimensions{
		Size: image.Point{
			X: gtx.Constraints.Max.X,
			Y: offsetY + scaledHeight,
		},
	}
}

// renderPlaceholder renders a placeholder for Scrappy Doo until actual sprites are available
// Task 32: This is the Scrappy Doo character sprite/asset placeholder
func (cr *CharacterRenderer) renderPlaceholder(gtx layout.Context) layout.Dimensions {
	// Define character size
	charWidth := gtx.Dp(unit.Dp(120))
	charHeight := gtx.Dp(unit.Dp(150))

	// Center the character
	offsetX := (gtx.Constraints.Max.X - charWidth) / 2
	offsetY := gtx.Dp(unit.Dp(20))

	// Create a simple placeholder representation of Scrappy Doo
	// This will be replaced with actual sprite rendering later

	// Draw body (brown/tan color for Scrappy Doo)
	bodyRect := image.Rectangle{
		Min: image.Point{X: offsetX + charWidth/4, Y: offsetY + charHeight/3},
		Max: image.Point{X: offsetX + 3*charWidth/4, Y: offsetY + charHeight},
	}
	paint.FillShape(gtx.Ops, color.NRGBA{R: 139, G: 90, B: 43, A: 255}, clip.Rect(bodyRect).Op())

	// Draw head (circle)
	headRadius := charWidth / 3
	headCenter := image.Point{
		X: offsetX + charWidth/2,
		Y: offsetY + headRadius,
	}
	headRect := image.Rectangle{
		Min: image.Point{X: headCenter.X - headRadius, Y: headCenter.Y - headRadius},
		Max: image.Point{X: headCenter.X + headRadius, Y: headCenter.Y + headRadius},
	}
	paint.FillShape(gtx.Ops, color.NRGBA{R: 139, G: 90, B: 43, A: 255}, clip.Ellipse{Min: headRect.Min, Max: headRect.Max}.Op(gtx.Ops))

	// Draw eyes
	eyeSize := gtx.Dp(unit.Dp(8))
	leftEye := image.Rectangle{
		Min: image.Point{X: headCenter.X - headRadius/2, Y: headCenter.Y - eyeSize/2},
		Max: image.Point{X: headCenter.X - headRadius/2 + eyeSize, Y: headCenter.Y + eyeSize/2},
	}
	rightEye := image.Rectangle{
		Min: image.Point{X: headCenter.X + headRadius/2 - eyeSize, Y: headCenter.Y - eyeSize/2},
		Max: image.Point{X: headCenter.X + headRadius/2, Y: headCenter.Y + eyeSize/2},
	}
	paint.FillShape(gtx.Ops, color.NRGBA{R: 0, G: 0, B: 0, A: 255}, clip.Rect(leftEye).Op())
	paint.FillShape(gtx.Ops, color.NRGBA{R: 0, G: 0, B: 0, A: 255}, clip.Rect(rightEye).Op())

	// Draw arms based on animation state
	cr.renderArms(gtx, offsetX, offsetY, charWidth, charHeight)

	// Draw legs
	legWidth := charWidth / 6
	leftLeg := image.Rectangle{
		Min: image.Point{X: offsetX + charWidth/3, Y: offsetY + 2*charHeight/3},
		Max: image.Point{X: offsetX + charWidth/3 + legWidth, Y: offsetY + charHeight},
	}
	rightLeg := image.Rectangle{
		Min: image.Point{X: offsetX + 2*charWidth/3 - legWidth, Y: offsetY + 2*charHeight/3},
		Max: image.Point{X: offsetX + 2*charWidth/3, Y: offsetY + charHeight},
	}
	paint.FillShape(gtx.Ops, color.NRGBA{R: 139, G: 90, B: 43, A: 255}, clip.Rect(leftLeg).Op())
	paint.FillShape(gtx.Ops, color.NRGBA{R: 139, G: 90, B: 43, A: 255}, clip.Rect(rightLeg).Op())

	return layout.Dimensions{
		Size: image.Point{
			X: gtx.Constraints.Max.X,
			Y: offsetY + charHeight,
		},
	}
}

// renderArms renders the arms based on the current animation state
func (cr *CharacterRenderer) renderArms(gtx layout.Context, offsetX, offsetY, charWidth, charHeight int) {
	state := cr.character.GetCurrentState()
	armWidth := gtx.Dp(unit.Dp(12))
	armLength := gtx.Dp(unit.Dp(40))

	// Base arm positions (at shoulders)
	leftShoulder := image.Point{X: offsetX + charWidth/4, Y: offsetY + charHeight/3}
	rightShoulder := image.Point{X: offsetX + 3*charWidth/4, Y: offsetY + charHeight/3}

	switch state {
	case AnimationStateJabLeft:
		// Jab animation: left arm extends forward
		// Calculate extension based on animation frame
		frame := cr.character.GetCurrentFrame()
		if frame != nil {
			anim := cr.character.GetAnimation(AnimationStateJabLeft)
			if anim != nil {
				currentFrame := cr.character.GetCurrentFrameIndex()
				if currentFrame < len(anim.Frames) {
					// Extend left arm forward during jab
					extension := float32(currentFrame) / float32(len(anim.Frames))
					if extension > 0.5 {
						extension = 1.0 - extension // Retract after peak
					} else {
						extension = extension * 2 // Extend to peak
					}

					leftHand := image.Point{
						X: leftShoulder.X + int(float32(armLength)*extension),
						Y: leftShoulder.Y,
					}

					// Draw extended left arm
					armRect := image.Rectangle{
						Min: leftShoulder,
						Max: image.Point{X: leftHand.X + armWidth, Y: leftHand.Y + armWidth},
					}
					paint.FillShape(gtx.Ops, color.NRGBA{R: 139, G: 90, B: 43, A: 255}, clip.Rect(armRect).Op())

					// Draw right arm in guard position
					rightHand := image.Point{
						X: rightShoulder.X - armLength/2,
						Y: rightShoulder.Y + armLength/2,
					}
					rightArmRect := image.Rectangle{
						Min: rightShoulder,
						Max: image.Point{X: rightHand.X + armWidth, Y: rightHand.Y + armWidth},
					}
					paint.FillShape(gtx.Ops, color.NRGBA{R: 139, G: 90, B: 43, A: 255}, clip.Rect(rightArmRect).Op())
					return
				}
			}
		}
		// Fall through to default if animation not ready
		fallthrough
	default:
		// Default/idle pose: arms in guard position
		leftHand := image.Point{
			X: leftShoulder.X + armLength/2,
			Y: leftShoulder.Y + armLength/2,
		}
		rightHand := image.Point{
			X: rightShoulder.X - armLength/2,
			Y: rightShoulder.Y + armLength/2,
		}

		// Draw left arm
		leftArmRect := image.Rectangle{
			Min: leftShoulder,
			Max: image.Point{X: leftHand.X + armWidth, Y: leftHand.Y + armWidth},
		}
		paint.FillShape(gtx.Ops, color.NRGBA{R: 139, G: 90, B: 43, A: 255}, clip.Rect(leftArmRect).Op())

		// Draw right arm
		rightArmRect := image.Rectangle{
			Min: rightShoulder,
			Max: image.Point{X: rightHand.X + armWidth, Y: rightHand.Y + armWidth},
		}
		paint.FillShape(gtx.Ops, color.NRGBA{R: 139, G: 90, B: 43, A: 255}, clip.Rect(rightArmRect).Op())
	}
}
