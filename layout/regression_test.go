package layout

import "testing"

// TestSplitMinExceedsAvailableNoUnderflow 验证 Min 超量后 remaining 饱和为 0，
// 后续 Length/Max 不会因 uint16 下溢而拿到错误的大尺寸。
func TestSplitMinExceedsAvailableNoUnderflow(t *testing.T) {
	areas := Vertical(NewMin(15), NewLength(5)).Split(Rect{Width: 10, Height: 10})
	if len(areas) != 2 {
		t.Fatalf("len(areas) = %d, want 2", len(areas))
	}
	// Min 可超量（保持既有语义），但 Length 应得 0 而非下溢的大值
	if areas[1].Height != 0 {
		t.Fatalf("Length area height = %d, want 0", areas[1].Height)
	}
}

// TestSplitFillNoMultiplyOverflow 验证 Fill 按权重分配用 32 位运算，
// remaining*value 不因 uint16 回绕导致比例错误。
func TestSplitFillNoMultiplyOverflow(t *testing.T) {
	areas := Horizontal(NewFill(256), NewFill(256)).Split(Rect{Width: 256, Height: 1})
	if areas[0].Width != 128 || areas[1].Width != 128 {
		t.Fatalf("fill widths = [%d, %d], want [128, 128]", areas[0].Width, areas[1].Width)
	}
}

// TestCenteredOversizedNoUnderflow 验证 Centered 目标尺寸大于容器时不再回绕。
func TestCenteredOversizedNoUnderflow(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 10, Height: 10}.Centered(20, 20)
	if r.X != 0 || r.Y != 0 {
		t.Fatalf("Centered oversized = (%d, %d), want (0, 0)", r.X, r.Y)
	}
}

// TestOffsetSaturatesUpperBound 验证 Offset 正向溢出饱和到 uint16 最大值。
func TestOffsetSaturatesUpperBound(t *testing.T) {
	r := Rect{X: 65000, Y: 0, Width: 5, Height: 1}.Offset(1000, 0)
	if r.X != 65535 {
		t.Fatalf("Offset X = %d, want 65535 (saturated)", r.X)
	}
}