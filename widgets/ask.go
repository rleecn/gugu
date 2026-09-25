package widgets

import (
	"strings"
	"unicode/utf8"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
)

// ---------------------------------------------------------------------------
// Ask：适用于 AI 应用的选择卡片。
//
// 把"必须由用户拍板"的分叉点交回用户——
// 展示一个问题与若干候选项（推荐 2-4 个），用户在选项间选择（单选/多选），
// 也可以随时直接键入自定义答案。模型给出的选项未必覆盖用户真实意图，
// 因此提交结果不保证来自选项列表，调用方不得假定 Answers() 一定等于某个选项。
//
// 组件只负责渲染与按键语义；何时渲染、提交后结果如何回填由宿主应用决定。
// ---------------------------------------------------------------------------

// AskEvent 表示一次按键处理后卡片的状态迁移结果。
type AskEvent int

const (
	// AskEventNone 继续等待用户操作。
	AskEventNone AskEvent = iota
	// AskEventSubmit 用户已确认，调用 Ask.Answers 读取作答。
	AskEventSubmit
	// AskEventDismiss 用户主动跳过（Esc），语义为"别替我做决定"，
	// 宿主不应据此替用户选定任何默认项继续推进。
	AskEventDismiss
)

const (
	askHintSingle = "↑↓ 选择  Enter 确认  Tab 输入  Esc 跳过"
	askHintMulti  = "↑↓ 选择  空格 勾选  Tab 输入  Enter 确认  Esc 跳过"
)

// AskState 管理 Ask 卡片的交互状态，与组件本体分离以便外部持有。
type AskState struct {
	cursor    int    // 高亮选项下标（单选模式下同时是选中项）
	checked   []bool // 多选模式各选项的勾选状态
	inputMode bool   // 焦点是否在自定义输入行
	input     Input  // 复用 Input 的编辑状态（值/光标/选区/横向滚动）
	lastPress int    // 最近一次点击命中的选项下标，-1 表示无（用于"再次点击同一项=确认"）
}

// NewAskState creates an AskState for the given number of options.
func NewAskState(optionCount int) AskState {
	optionCount = max(optionCount, 0)
	st := AskState{input: NewInput().SetBlock(NoBlock()), lastPress: -1}
	if optionCount > 0 {
		st.checked = make([]bool, optionCount)
	}
	return st
}

// IsState implements terminal.State marker interface.
func (s *AskState) IsState() {}

// Cursor returns the highlighted option index.
func (s *AskState) Cursor() int { return s.cursor }

// SetCursor sets the highlighted option index (clamped to >= 0;
// upper bound is enforced by Ask at render/key handling time).
func (s *AskState) SetCursor(i int) {
	s.cursor = max(i, 0)
}

// Checked reports whether option i is checked (multi mode).
func (s *AskState) Checked(i int) bool {
	return i >= 0 && i < len(s.checked) && s.checked[i]
}

// InInputMode reports whether the custom-input row has focus.
func (s *AskState) InInputMode() bool { return s.inputMode }

// InputValue returns the current custom input text.
func (s *AskState) InputValue() string { return s.input.Value() }

// Ask 是一个带问题、选项与自定义输入行的选择卡片，支持单选与多选。
type Ask struct {
	block             Block
	question          string
	questionStyle     style.Style
	options           []string
	multi             bool
	style             style.Style
	highlightStyle    style.Style
	hintStyle         style.Style
	checkedSymbol     string
	uncheckedSymbol   string
	customPlaceholder string
	customMaxLength   int
	hintText          string
	hintSet           bool
	state             AskState // 无状态 Render 使用；推荐通过 RenderStateful 外部管理
}

// NewAsk creates an Ask card with the given question and option labels.
// 默认单选模式；多选请调用 SetMulti(true)。
func NewAsk(question string, options []string) Ask {
	return Ask{
		block:             NewBlock().SetBorders(BorderAll), // 默认圆角边框卡片
		question:          question,
		questionStyle:     style.NewStyle(),
		options:           options,
		style:             style.NewStyle(),
		highlightStyle:    style.NewStyle().SetBg(style.Cyan).SetFg(style.Black),
		hintStyle:         style.NewStyle().SetFg(style.DarkGray),
		checkedSymbol:     "◉",
		uncheckedSymbol:   "○",
		customPlaceholder: "✎ 自定义输入…",
		state:             NewAskState(len(options)),
	}
}

// SetBlock sets the wrapping block（标题即卡片 header，可用 SetTitle 设置）.
func (a Ask) SetBlock(b Block) Ask {
	a.block = b
	return a
}

// SetMulti enables multi-select mode（空格勾选，Enter 提交全部勾选项）.
func (a Ask) SetMulti(on bool) Ask {
	a.multi = on
	return a
}

// SetStyle sets the base style of the card.
func (a Ask) SetStyle(s style.Style) Ask {
	a.style = s
	return a
}

// SetQuestionStyle sets the question text style.
func (a Ask) SetQuestionStyle(s style.Style) Ask {
	a.questionStyle = s
	return a
}

// SetHighlightStyle sets the style of the cursor row / focused input row.
func (a Ask) SetHighlightStyle(s style.Style) Ask {
	a.highlightStyle = s
	return a
}

// SetHintStyle sets the style of the bottom key-hint line.
func (a Ask) SetHintStyle(s style.Style) Ask {
	a.hintStyle = s
	return a
}

// SetHint sets the key-hint line; empty string hides it.
// 默认按单选/多选模式自动生成提示。
func (a Ask) SetHint(s string) Ask {
	a.hintText = s
	a.hintSet = true
	return a
}

// SetSymbols sets the marks shown before each option.
func (a Ask) SetSymbols(checked, unchecked string) Ask {
	a.checkedSymbol = checked
	a.uncheckedSymbol = unchecked
	return a
}

// SetCustomPlaceholder sets the placeholder of the custom-input row.
func (a Ask) SetCustomPlaceholder(s string) Ask {
	a.customPlaceholder = s
	return a
}

// SetCustomMaxLength limits the number of runes accepted by the custom input.
// 0 means unlimited。建议宿主设置上限：答案会回填进模型上下文，
// 不设上限时一次误粘贴就可能撑爆上下文。
func (a Ask) SetCustomMaxLength(n int) Ask {
	a.customMaxLength = n
	return a
}

// Len returns the number of options（用于创建配套的 AskState）.
func (a Ask) Len() int { return len(a.options) }

// State returns a copy of the internal state（无状态 Render 时可读取）.
func (a Ask) State() AskState { return a.state }

// SetState replaces the internal state.
func (a Ask) SetState(s AskState) Ask {
	a.state = s
	return a
}

// Answers 返回当前作答。自定义输入行有内容时仅返回该文本；
// 否则单选返回高亮项、多选返回全部勾选项。结果可能为空（尚未作答/跳过）。
func (a Ask) Answers(st *AskState) []string {
	if st == nil {
		return nil
	}
	if st.inputMode {
		if v := strings.TrimSpace(st.input.Value()); v != "" {
			return []string{v}
		}
	}
	if a.multi {
		out := make([]string, 0, len(a.options))
		for i, c := range st.checked {
			if c && i < len(a.options) {
				out = append(out, a.options[i])
			}
		}
		return out
	}
	if len(a.options) == 0 {
		return nil
	}
	i := st.cursor
	if i >= len(a.options) {
		i = len(a.options) - 1
	}
	return []string{a.options[i]}
}

// HandleKey 处理一个按键事件并迁移卡片状态，返回事件结果。
//
// 按键语义：
//   - ↑/↓：在选项间移动；越过最后一项进入自定义输入行，输入行按 ↑ 返回选项
//   - Tab：选项行与自定义输入行之间切换焦点
//   - 空格：多选模式下勾选/取消高亮项（单选模式无操作，高亮即选中）
//   - 可见字符：直接进入自定义输入行并插入（Ctrl/Alt 组合键除外）
//   - Enter：确认。输入行有内容时提交自定义答案；输入行为空时返回选项区；
//     多选模式下未勾选任何项时 Enter 不生效（主动跳过请用 Esc）
//   - Esc：跳过本次提问（AskEventDismiss）
func (a Ask) HandleKey(ev terminal.KeyEvent, st *AskState) AskEvent {
	if st == nil || ev.Release {
		return AskEventNone
	}
	switch ev.Code {
	case terminal.KeyEsc:
		return AskEventDismiss
	case terminal.KeyUp:
		a.moveCursor(st, -1)
		return AskEventNone
	case terminal.KeyDown:
		a.moveCursor(st, 1)
		return AskEventNone
	case terminal.KeyTab:
		st.inputMode = !st.inputMode
		return AskEventNone
	case terminal.KeyEnter:
		return a.submit(st)
	}
	if st.inputMode {
		switch ev.Code {
		case terminal.KeyBackspace:
			st.input.DeleteCharBack()
		case terminal.KeyDelete:
			st.input.DeleteCharForward()
		case terminal.KeyLeft:
			st.input.MoveCursorLeft()
		case terminal.KeyRight:
			st.input.MoveCursorRight()
		case terminal.KeyHome:
			st.input.MoveCursorHome()
		case terminal.KeyEnd:
			st.input.MoveCursorEnd()
		default:
			a.insertText(st, ev)
		}
		return AskEventNone
	}
	// 选项区：空格在多选模式下切换勾选；其余可见字符直接进入自定义输入
	if ev.IsChar() && ev.Text == " " && a.multi && st.cursor < len(st.checked) &&
		!ev.Modifiers.HasCtrl() && !ev.Modifiers.HasAlt() {
		st.checked[st.cursor] = !st.checked[st.cursor]
		return AskEventNone
	}
	a.insertText(st, ev)
	return AskEventNone
}

// HandleMouse 处理鼠标事件，返回值语义与 HandleKey 一致。area 必须与当前
// 渲染该卡片的区域一致（命中检测与渲染共用 computeRows 的几何推导）。
//
// 鼠标语义：
//   - 左键点击选项行：单选选中该选项，再次点击同一选项 = 确认提交；
//     多选切换勾选（不触发提交，确认仍走 Enter）
//   - 左键点击输入行：聚焦自定义输入
//   - 滚轮：上下移动选项光标（与 ↑/↓ 等价）
//   - 其余按键动作（右键/中键/移动/悬停）与点击到卡片外均忽略
func (a Ask) HandleMouse(ev terminal.MouseEvent, st *AskState, area layout.Rect) AskEvent {
	if st == nil {
		return AskEventNone
	}
	switch ev.Action {
	case terminal.MousePress:
	case terminal.MouseWheelUp:
		a.moveCursor(st, -1)
		return AskEventNone
	case terminal.MouseWheelDown:
		a.moveCursor(st, 1)
		return AskEventNone
	default:
		return AskEventNone
	}

	r := a.computeRows(area)
	if r.inner.IsEmpty() {
		return AskEventNone
	}
	x, y := int(ev.X), int(ev.Y)
	if x < int(r.inner.X) || x >= int(r.inner.Right()) {
		return AskEventNone
	}
	bottom := int(r.inner.Bottom())

	// 选项行：行号可能因区域裁剪而不完整渲染，命中范围取 inner.Bottom 上界
	if y >= r.optionsY && y < min(r.optionsY+len(a.options), bottom) {
		i := y - r.optionsY
		st.inputMode = false
		if a.multi {
			st.cursor = i
			if i < len(st.checked) {
				st.checked[i] = !st.checked[i]
			}
		} else if st.cursor == i && st.lastPress == i {
			return a.submit(st) // 再次点击已选中项 = 确认
		} else {
			st.cursor = i
		}
		st.lastPress = i
		return AskEventNone
	}
	// 输入行（紧凑布局下紧邻选项行，无空行间隔，命中条件不依赖 gaps）
	if y == r.inputY && y < bottom {
		st.inputMode = true
	}
	return AskEventNone
}

// insertText 在自定义输入行插入字符事件携带的文本。
// 仅接受无 Ctrl/Alt/Super 修饰的输入，避免 Ctrl+C 等组合键被当作答案内容；
// maxLength 在插入路径强制执行（渲染期临时拷贝上的配置不参与编辑）。
func (a Ask) insertText(st *AskState, ev terminal.KeyEvent) {
	if !ev.IsChar() || ev.Modifiers.HasCtrl() || ev.Modifiers.HasAlt() || ev.Super {
		return
	}
	if !st.inputMode {
		st.inputMode = true
	}
	text := ev.Text
	if a.customMaxLength > 0 {
		remaining := a.customMaxLength - utf8.RuneCountInString(st.input.Value())
		if remaining <= 0 {
			return
		}
		if runes := []rune(text); len(runes) > remaining {
			text = string(runes[:remaining])
		}
	}
	st.input.InsertString(text)
}

// moveCursor 在选项间移动光标，越过边界时与输入行互转焦点。
func (a Ask) moveCursor(st *AskState, delta int) {
	n := len(a.options)
	if st.inputMode {
		if delta < 0 && n > 0 {
			st.inputMode = false
			st.cursor = n - 1
		}
		return
	}
	if n == 0 {
		st.inputMode = true
		return
	}
	st.cursor += delta
	if st.cursor < 0 {
		st.cursor = 0
	}
	if st.cursor >= n {
		st.cursor = n - 1
		st.inputMode = true
	}
}

// submit 处理 Enter：返回提交事件，或在没有可提交内容时迁移状态。
func (a Ask) submit(st *AskState) AskEvent {
	if st.inputMode {
		if strings.TrimSpace(st.input.Value()) != "" {
			return AskEventSubmit
		}
		st.inputMode = false // 空输入回车 = 回到选项区
		return AskEventNone
	}
	if len(a.options) == 0 {
		return AskEventNone
	}
	if a.multi && !a.checkedAny(st) {
		return AskEventNone // 防止误触空提交；主动放弃请用 Esc
	}
	return AskEventSubmit
}

func (a Ask) checkedAny(st *AskState) bool {
	for _, c := range st.checked {
		if c {
			return true
		}
	}
	return false
}

// Render renders the card using its internal state.
func (a Ask) Render(area layout.Rect, buf *buffer.Buffer) {
	a.renderWithState(area, buf, &a.state)
}

// RenderStateful implements terminal.StatefulWidget.
func (a Ask) RenderStateful(area layout.Rect, buf *buffer.Buffer, state terminal.State) {
	if s, ok := state.(*AskState); ok {
		a.renderWithState(area, buf, s)
	}
}

func (a Ask) renderWithState(area layout.Rect, buf *buffer.Buffer, st *AskState) {
	a.block.Render(area, buf)
	r := a.computeRows(area)
	if r.inner.IsEmpty() {
		return
	}

	// 基础样式铺满内部区域
	for y := r.inner.Y; y < r.inner.Bottom(); y++ {
		for x := r.inner.X; x < r.inner.Right(); x++ {
			if cell := buf.CellAt(x, y); cell != nil {
				cell.SetStyle(a.style)
			}
		}
	}

	if r.questionH > 0 {
		a.questionPara().Render(
			layout.Rect{X: r.inner.X, Y: uint16(r.questionY), Width: r.inner.Width, Height: uint16(r.questionH)}, buf)
	}
	row := uint16(r.optionsY)
	for i, opt := range a.options {
		if row >= r.inner.Bottom() {
			break
		}
		a.renderOption(buf, r.inner, row, i, opt, st)
		row++
	}
	if r.gaps == 2 && row < r.inner.Bottom() {
		row++
	}
	if row < r.inner.Bottom() {
		a.renderInputRow(buf, r.inner, row, st)
		row++
	}
	if r.hintH == 1 && row < r.inner.Bottom() {
		buf.SetStringn(r.inner.X, row, a.hint(), r.inner.Width, a.hintStyle)
	}
}

// askRows 描述卡片在给定区域内的行布局，渲染与鼠标命中检测共用同一套几何。
type askRows struct {
	inner     layout.Rect
	questionY int
	questionH int
	optionsY  int // 首个选项行（绝对坐标）
	inputY    int // 输入行（绝对坐标，未裁剪值；命中检测需再对 inner.Bottom 取界）
	gaps      int
	hintH     int
}

// computeRows 推导内部布局：问题(0..questionH) + 空行 + 选项 + 空行 + 输入行 + 提示行；
// 高度不足时依次折叠空行与问题行（与渲染折叠顺序一致）。
func (a Ask) computeRows(area layout.Rect) askRows {
	r := askRows{inner: a.block.Inner(area)}
	// 问题段落渲染在 inner 坐标系（宽度/高度均取自 inner），
	// Y 也必须取 inner.Y：取 area.Y 会在带边框时叠画上边框/标题行
	r.questionY = int(r.inner.Y)
	if r.inner.IsEmpty() {
		return r
	}
	hint := a.hint()
	if hint != "" {
		r.hintH = 1
	}
	fixed := len(a.options) + 1 + r.hintH // 选项 + 输入行 + 提示行
	r.gaps = 2                            // 问题与选项、选项与输入行之间的空行
	if height := int(r.inner.Height); height < fixed+r.gaps {
		r.gaps = max(height-fixed, 0)
	}
	r.questionH = max(int(r.inner.Height)-fixed-r.gaps, 0)

	r.optionsY = int(r.inner.Y) + r.questionH
	if r.gaps > 0 {
		r.optionsY++
	}
	r.inputY = r.optionsY + len(a.options)
	if r.gaps == 2 {
		r.inputY++
	}
	return r
}

// questionPara 构造内部问题段落（无边框、按词换行）。
func (a Ask) questionPara() Paragraph {
	return NewParagraph(a.question).
		SetBlock(NoBlock()).
		SetWrap(WrapWord).
		SetStyle(a.questionStyle)
}

func (a Ask) renderOption(buf *buffer.Buffer, inner layout.Rect, row uint16, i int, label string, st *AskState) {
	rowStyle := a.style
	if i == st.cursor && !st.inputMode {
		rowStyle = a.highlightStyle
	}
	for x := inner.X; x < inner.Right(); x++ {
		if cell := buf.CellAt(x, row); cell != nil {
			cell.SetStyle(rowStyle)
		}
	}
	checked := i == st.cursor // 单选：高亮即选中
	if a.multi {
		checked = st.Checked(i)
	}
	mark := a.uncheckedSymbol
	if checked {
		mark = a.checkedSymbol
	}
	buf.SetStringn(inner.X, row, mark+" "+label, inner.Width, rowStyle)
}

func (a Ask) renderInputRow(buf *buffer.Buffer, inner layout.Rect, row uint16, st *AskState) {
	rowStyle := a.style
	if st.inputMode {
		rowStyle = a.highlightStyle
		for x := inner.X; x < inner.Right(); x++ {
			if cell := buf.CellAt(x, row); cell != nil {
				cell.SetStyle(rowStyle)
			}
		}
	}
	in := st.input.
		SetPlaceholder(a.customPlaceholder).
		SetStyle(rowStyle).
		SetFocused(st.inputMode)
	in.Render(layout.Rect{X: inner.X, Y: row, Width: inner.Width, Height: 1}, buf)
}

// hint 返回当前应显示的提示文案。
func (a Ask) hint() string {
	if a.hintSet {
		return a.hintText
	}
	if a.multi {
		return askHintMulti
	}
	return askHintSingle
}
