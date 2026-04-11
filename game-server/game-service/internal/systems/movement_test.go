package systems

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// PlayerRadius = 20

func TestSweptX(t *testing.T) {
	tests := []struct {
		name                       string
		playerX, playerY           float64
		wallX, wallY, wallW, wallH float64
		dx                         float64
		want                       float64
	}{
		{
			name:    "hit wall moving right",
			playerX: 100, playerY: 300, // player right edge = 120
			wallX: 140, wallY: 280, wallW: 20, wallH: 40, // wall left = 140
			dx:   50,
			want: 0.4, // (140 - 120) / 50 = 0.4
		},
		{
			name:    "hit wall moving left",
			playerX: 200, playerY: 300, // player left edge = 180
			wallX: 140, wallY: 280, wallW: 20, wallH: 40, // wall right = 160
			dx:   -50,
			want: 0.4, // (160 - 180) / -50 = 0.4
		},
		{
			name:    "wall too far, no collision",
			playerX: 100, playerY: 300, // player right edge = 120
			wallX: 400, wallY: 280, wallW: 20, wallH: 40,
			dx:   50,
			want: 1, // (400 - 120) / 50 = 5.6 > 1
		},
		{
			name:    "moving away from wall",
			playerX: 100, playerY: 300,
			wallX: 140, wallY: 280, wallW: 20, wallH: 40,
			dx:   -50, // moving left, wall is on right
			want: 1,   // tEnter negative
		},
		{
			name:    "dx is zero",
			playerX: 100, playerY: 300,
			wallX: 140, wallY: 280, wallW: 20, wallH: 40,
			dx:   0,
			want: 1,
		},
		{
			name:    "Y no overlap, wall above player",
			playerX: 100, playerY: 350, // player top = 330
			wallX: 140, wallY: 280, wallW: 20, wallH: 40, // wall bottom = 320
			dx:   50,
			want: 1, // playerTop(330) >= wallBottom(320)
		},
		{
			name:    "Y no overlap, wall below player",
			playerX: 100, playerY: 250, // player bottom = 270
			wallX: 140, wallY: 280, wallW: 20, wallH: 40, // wall top = 280
			dx:   50,
			want: 1, // playerBottom(270) <= wallTop(280)
		},
		{
			name:    "Y exactly touching, no collision",
			playerX: 100, playerY: 260, // player bottom = 280
			wallX: 140, wallY: 280, wallW: 20, wallH: 40, // wall top = 280
			dx:   50,
			want: 1, // playerBottom(280) <= wallTop(280)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SweptX(tt.playerX, tt.playerY, tt.wallX, tt.wallY, tt.wallW, tt.wallH, tt.dx)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

func TestSweptY(t *testing.T) {
	tests := []struct {
		name                       string
		playerX, playerY           float64
		wallX, wallY, wallW, wallH float64
		dy                         float64
		want                       float64
	}{
		{
			name:    "hit wall moving down",
			playerX: 300, playerY: 100, // player bottom edge = 120
			wallX: 280, wallY: 140, wallW: 40, wallH: 20, // wall top = 140
			dy:   50,
			want: 0.4, // (140 - 120) / 50 = 0.4
		},
		{
			name:    "hit wall moving up",
			playerX: 300, playerY: 200, // player top edge = 180
			wallX: 280, wallY: 140, wallW: 40, wallH: 20, // wall bottom = 160
			dy:   -50,
			want: 0.4, // (160 - 180) / -50 = 0.4
		},
		{
			name:    "wall too far, no collision",
			playerX: 300, playerY: 100,
			wallX: 280, wallY: 400, wallW: 40, wallH: 20,
			dy:   50,
			want: 1,
		},
		{
			name:    "moving away from wall",
			playerX: 300, playerY: 100,
			wallX: 280, wallY: 140, wallW: 40, wallH: 20,
			dy:   -50,
			want: 1,
		},
		{
			name:    "dy is zero",
			playerX: 300, playerY: 100,
			wallX: 280, wallY: 140, wallW: 40, wallH: 20,
			dy:   0,
			want: 1,
		},
		{
			name:    "X no overlap, wall to the right",
			playerX: 250, playerY: 100, // player right = 270
			wallX: 280, wallY: 140, wallW: 40, wallH: 20, // wall left = 280
			dy:   50,
			want: 1, // playerRight(270) <= wallLeft(280)
		},
		{
			name:    "X no overlap, wall to the left",
			playerX: 350, playerY: 100, // player left = 330
			wallX: 280, wallY: 140, wallW: 40, wallH: 20, // wall right = 320
			dy:   50,
			want: 1, // playerLeft(330) >= wallRight(320)
		},
		{
			name:    "X exactly touching, no collision",
			playerX: 260, playerY: 100, // player right = 280
			wallX: 280, wallY: 140, wallW: 40, wallH: 20, // wall left = 280
			dy:   50,
			want: 1, // playerRight(280) <= wallLeft(280)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SweptY(tt.playerX, tt.playerY, tt.wallX, tt.wallY, tt.wallW, tt.wallH, tt.dy)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}
