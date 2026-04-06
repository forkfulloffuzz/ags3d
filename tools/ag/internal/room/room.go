// Package room provides types and a parser for .agroom config files.
//
// A .agroom file describes the gameplay nodes of a single room:
// cameras, walkable surfaces, blocker volumes, spawn points, hotspots,
// and named waypoints. It does not contain mesh or lighting data.
//
// Example:
//
//	Room "start" {
//	    initial_camera = "main"
//
//	    Camera "main" {
//	        position = (4.79, 5.52, 5.60)
//	        look_at  = (0.0, 0.0, 0.0)
//	    }
//
//	    Point "door_left" {
//	        position = (3.12, 0.18, 3.43)
//	    }
//
//	    WalkableSurface "floor" {
//	        size   = (10.0, 10.0)
//	        offset = (0.0, -0.05, 0.0)
//	    }
//
//	    BlockerVolume "wall" {
//	        size     = (1.0, 2.0, 3.0)
//	        position = (0.0, 1.0, 1.15)
//	    }
//
//	    SpawnPoint "player_start" {
//	        character = "player"
//	        position  = (-4.0, 0.0, -3.0)
//	    }
//
//	    Hotspot "bookshelf" {
//	        size     = (1.0, 2.0, 0.3)
//	        position = (2.0, 1.0, -4.8)
//	    }
//	}
package room

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// --------------------------------------------------------------------------
// Data types
// --------------------------------------------------------------------------

// Vec3 is a 3-component position or size vector (X, Y, Z).
type Vec3 struct{ X, Y, Z float64 }

// Vec2 is a 2-component XZ-plane size (used for WalkableSurface.Size).
type Vec2 struct{ X, Z float64 }

// RoomDialogue holds the parsed content of a `dialogue { ... }` block inside
// a .agroom file.
type RoomDialogue struct {
	OnEnter       string // node title to trigger on first room enter
	OnEnterRepeat string // node title to trigger on subsequent room enters
}

// RoomData is the fully parsed representation of one .agroom file.
type RoomData struct {
	Name             string
	InitialCamera    string
	Cameras          []CameraData
	Points           []PointData
	WalkableSurfaces []WalkableSurfaceData
	BlockerVolumes   []BlockerVolumeData
	SpawnPoints      []SpawnPointData
	Hotspots         []HotspotData
	TriggerRegions   []TriggerRegionData
	Dialogue         *RoomDialogue // nil if no dialogue block present
}

// ValidateDialogueRefs checks that all node titles listed in the dialogue block
// exist in knownTitles. Returns one error per unknown title.
func (rd *RoomData) ValidateDialogueRefs(knownTitles map[string]bool) []error {
	if rd.Dialogue == nil {
		return nil
	}
	var errs []error
	check := func(field, title string) {
		if title != "" && !knownTitles[title] {
			errs = append(errs, fmt.Errorf("%s: dialogue.%s: node title %q not found", rd.Name, field, title))
		}
	}
	check("on_enter", rd.Dialogue.OnEnter)
	check("on_enter_repeat", rd.Dialogue.OnEnterRepeat)
	return errs
}

// TriggerRegionData holds the parsed data for a TriggerRegion block.
type TriggerRegionData struct {
	Name     string
	Size     Vec3
	Position Vec3
}

// CameraData holds the parsed data for a Camera block.
type CameraData struct {
	Name        string
	Position    Vec3 // explicit position; only valid when HasPosition is true
	LookAt      Vec3 // explicit look_at; only valid when HasLookAt is true
	HasPosition bool // true when position was specified in the .agroom file
	HasLookAt   bool // true when look_at was specified in the .agroom file
}

// PointData holds the parsed data for a Point (named waypoint) block.
type PointData struct {
	Name     string
	Position Vec3
}

// WalkableSurfaceData holds the parsed data for a WalkableSurface block.
type WalkableSurfaceData struct {
	Name   string
	Size   Vec2 // XZ-plane size; Y is always 0 (flat plane)
	Offset Vec3 // world-space positional offset
}

// BlockerVolumeData holds the parsed data for a BlockerVolume block.
type BlockerVolumeData struct {
	Name     string
	Size     Vec3
	Position Vec3
}

// SpawnPointData holds the parsed data for a SpawnPoint block.
type SpawnPointData struct {
	Name      string
	Character string // .agchar name (without extension)
	Position  Vec3
}

// HotspotData holds the parsed data for a Hotspot block.
type HotspotData struct {
	Name     string
	Size     Vec3
	Position Vec3
}

// --------------------------------------------------------------------------
// ParseError
// --------------------------------------------------------------------------

// ParseError is a parse error with file and line context.
type ParseError struct {
	File    string
	Line    int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Message)
}

// --------------------------------------------------------------------------
// Public entry point
// --------------------------------------------------------------------------

// ParseRoom parses src as a .agroom file and returns the RoomData.
// filename is used only in error messages.
func ParseRoom(filename, src string) (*RoomData, error) {
	p := &agparser{src: []rune(src), file: filename, line: 1}
	return p.parseRoom()
}

// --------------------------------------------------------------------------
// Parser
// --------------------------------------------------------------------------

type agparser struct {
	src  []rune
	pos  int
	file string
	line int
}

func (p *agparser) errorf(format string, args ...any) error {
	return &ParseError{File: p.file, Line: p.line, Message: fmt.Sprintf(format, args...)}
}

func (p *agparser) eof() bool { return p.pos >= len(p.src) }

func (p *agparser) peek() rune {
	if p.eof() {
		return 0
	}
	return p.src[p.pos]
}

func (p *agparser) advance() rune {
	ch := p.src[p.pos]
	p.pos++
	if ch == '\n' {
		p.line++
	}
	return ch
}

// skipWS skips whitespace and // line comments.
func (p *agparser) skipWS() {
	for !p.eof() {
		ch := p.peek()
		switch {
		case ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n':
			p.advance()
		case ch == '/' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '/':
			for !p.eof() && p.peek() != '\n' {
				p.advance()
			}
		default:
			return
		}
	}
}

// ident reads a sequence of letters, digits, and underscores.
func (p *agparser) ident() (string, error) {
	p.skipWS()
	if p.eof() || (!unicode.IsLetter(p.peek()) && p.peek() != '_') {
		if p.eof() {
			return "", p.errorf("unexpected end of file, expected identifier")
		}
		return "", p.errorf("expected identifier, got %q", p.peek())
	}
	var b strings.Builder
	for !p.eof() && (unicode.IsLetter(p.peek()) || unicode.IsDigit(p.peek()) || p.peek() == '_') {
		b.WriteRune(p.advance())
	}
	return b.String(), nil
}

// str reads a double-quoted string literal.
func (p *agparser) str() (string, error) {
	p.skipWS()
	if p.eof() || p.peek() != '"' {
		if p.eof() {
			return "", p.errorf("unexpected end of file, expected string")
		}
		return "", p.errorf("expected string literal, got %q", p.peek())
	}
	p.advance() // opening "
	var b strings.Builder
	for {
		if p.eof() {
			return "", p.errorf("unterminated string literal")
		}
		ch := p.advance()
		if ch == '"' {
			break
		}
		if ch == '\\' {
			if p.eof() {
				return "", p.errorf("unterminated escape sequence in string")
			}
			switch p.advance() {
			case 'n':
				b.WriteRune('\n')
			case 't':
				b.WriteRune('\t')
			case '"':
				b.WriteRune('"')
			case '\\':
				b.WriteRune('\\')
			default:
				return "", p.errorf("unknown escape sequence")
			}
		} else {
			b.WriteRune(ch)
		}
	}
	return b.String(), nil
}

// expect consumes a specific single character or returns an error.
func (p *agparser) expect(ch rune) error {
	p.skipWS()
	if p.eof() {
		return p.errorf("unexpected end of file, expected %q", ch)
	}
	if p.peek() != ch {
		return p.errorf("expected %q, got %q", ch, p.peek())
	}
	p.advance()
	return nil
}

// float reads a signed decimal number (no scientific notation).
func (p *agparser) float() (float64, error) {
	p.skipWS()
	start := p.pos
	if !p.eof() && (p.peek() == '-' || p.peek() == '+') {
		p.advance()
	}
	if p.eof() || !unicode.IsDigit(p.peek()) {
		return 0, p.errorf("expected number")
	}
	for !p.eof() && unicode.IsDigit(p.peek()) {
		p.advance()
	}
	if !p.eof() && p.peek() == '.' {
		p.advance()
		for !p.eof() && unicode.IsDigit(p.peek()) {
			p.advance()
		}
	}
	raw := string(p.src[start:p.pos])
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, p.errorf("invalid number %q", raw)
	}
	return v, nil
}

// tuple parses (f1, f2) or (f1, f2, f3) and returns the float slice.
func (p *agparser) tuple() ([]float64, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	var floats []float64
	for {
		f, err := p.float()
		if err != nil {
			return nil, err
		}
		floats = append(floats, f)
		p.skipWS()
		if p.eof() {
			return nil, p.errorf("unterminated tuple, expected ',' or ')'")
		}
		if p.peek() == ')' {
			p.advance()
			break
		}
		if p.peek() == ',' {
			p.advance()
			continue
		}
		return nil, p.errorf("expected ',' or ')' in tuple, got %q", p.peek())
	}
	if len(floats) < 2 || len(floats) > 3 {
		return nil, p.errorf("tuple must have 2 or 3 components, got %d", len(floats))
	}
	return floats, nil
}

func asVec3(fs []float64, field string) (Vec3, error) {
	if len(fs) != 3 {
		return Vec3{}, fmt.Errorf("%s requires 3 components, got %d", field, len(fs))
	}
	return Vec3{fs[0], fs[1], fs[2]}, nil
}

func asVec2(fs []float64, field string) (Vec2, error) {
	if len(fs) != 2 {
		return Vec2{}, fmt.Errorf("%s requires 2 components, got %d", field, len(fs))
	}
	return Vec2{fs[0], fs[1]}, nil
}

// --------------------------------------------------------------------------
// Top-level Room block
// --------------------------------------------------------------------------

func (p *agparser) parseRoom() (*RoomData, error) {
	p.skipWS()
	kw, err := p.ident()
	if err != nil {
		return nil, err
	}
	if kw != "Room" {
		return nil, p.errorf("expected 'Room', got %q", kw)
	}
	name, err := p.str()
	if err != nil {
		return nil, err
	}
	if err := p.expect('{'); err != nil {
		return nil, err
	}

	rd := &RoomData{Name: name}

	for {
		p.skipWS()
		if p.eof() {
			return nil, p.errorf("unexpected end of file inside Room block — missing '}'")
		}
		if p.peek() == '}' {
			p.advance()
			break
		}

		tok, err := p.ident()
		if err != nil {
			return nil, err
		}

		p.skipWS()
		if tok == "dialogue" && p.peek() == '{' {
			// dialogue { ... } — unnamed block
			p.advance() // consume '{'
			dlg, err := p.parseRoomDialogue()
			if err != nil {
				return nil, err
			}
			rd.Dialogue = dlg
		} else if p.peek() == '=' {
			// key = value
			p.advance()
			switch tok {
			case "initial_camera":
				v, err := p.str()
				if err != nil {
					return nil, err
				}
				rd.InitialCamera = v
			default:
				return nil, p.errorf("unknown Room property %q", tok)
			}
		} else {
			// TypeName "name" { ... }
			blockName, err := p.str()
			if err != nil {
				return nil, err
			}
			if err := p.expect('{'); err != nil {
				return nil, err
			}
			switch tok {
			case "Camera":
				cd, err := p.parseCamera(blockName)
				if err != nil {
					return nil, err
				}
				rd.Cameras = append(rd.Cameras, cd)
			case "Point":
				pd, err := p.parsePoint(blockName)
				if err != nil {
					return nil, err
				}
				rd.Points = append(rd.Points, pd)
			case "WalkableSurface":
				wd, err := p.parseWalkableSurface(blockName)
				if err != nil {
					return nil, err
				}
				rd.WalkableSurfaces = append(rd.WalkableSurfaces, wd)
			case "BlockerVolume":
				bd, err := p.parseBlockerVolume(blockName)
				if err != nil {
					return nil, err
				}
				rd.BlockerVolumes = append(rd.BlockerVolumes, bd)
			case "SpawnPoint":
				sd, err := p.parseSpawnPoint(blockName)
				if err != nil {
					return nil, err
				}
				rd.SpawnPoints = append(rd.SpawnPoints, sd)
			case "Hotspot":
				hd, err := p.parseHotspot(blockName)
				if err != nil {
					return nil, err
				}
				rd.Hotspots = append(rd.Hotspots, hd)
			case "TriggerRegion":
				td, err := p.parseTriggerRegion(blockName)
				if err != nil {
					return nil, err
				}
				rd.TriggerRegions = append(rd.TriggerRegions, td)
			default:
				return nil, p.errorf("unknown block type %q", tok)
			}
		}
	}

	p.skipWS()
	if !p.eof() {
		return nil, p.errorf("unexpected content after Room block")
	}
	return rd, nil
}

// --------------------------------------------------------------------------
// Block parsers
// --------------------------------------------------------------------------

func (p *agparser) parseCamera(name string) (CameraData, error) {
	cd := CameraData{Name: name}
	for {
		p.skipWS()
		if p.eof() {
			return cd, p.errorf("unterminated Camera %q block", name)
		}
		if p.peek() == '}' {
			p.advance()
			break
		}
		key, err := p.ident()
		if err != nil {
			return cd, err
		}
		if err := p.expect('='); err != nil {
			return cd, err
		}
		switch key {
		case "position":
			fs, err := p.tuple()
			if err != nil {
				return cd, err
			}
			if cd.Position, err = asVec3(fs, "position"); err != nil {
				return cd, p.errorf("%v", err)
			}
			cd.HasPosition = true
		case "look_at":
			fs, err := p.tuple()
			if err != nil {
				return cd, err
			}
			if cd.LookAt, err = asVec3(fs, "look_at"); err != nil {
				return cd, p.errorf("%v", err)
			}
			cd.HasLookAt = true
		default:
			return cd, p.errorf("unknown Camera property %q", key)
		}
	}
	return cd, nil
}

func (p *agparser) parsePoint(name string) (PointData, error) {
	pd := PointData{Name: name}
	for {
		p.skipWS()
		if p.eof() {
			return pd, p.errorf("unterminated Point %q block", name)
		}
		if p.peek() == '}' {
			p.advance()
			break
		}
		key, err := p.ident()
		if err != nil {
			return pd, err
		}
		if err := p.expect('='); err != nil {
			return pd, err
		}
		switch key {
		case "position":
			fs, err := p.tuple()
			if err != nil {
				return pd, err
			}
			if pd.Position, err = asVec3(fs, "position"); err != nil {
				return pd, p.errorf("%v", err)
			}
		default:
			return pd, p.errorf("unknown Point property %q", key)
		}
	}
	return pd, nil
}

func (p *agparser) parseWalkableSurface(name string) (WalkableSurfaceData, error) {
	wd := WalkableSurfaceData{Name: name}
	for {
		p.skipWS()
		if p.eof() {
			return wd, p.errorf("unterminated WalkableSurface %q block", name)
		}
		if p.peek() == '}' {
			p.advance()
			break
		}
		key, err := p.ident()
		if err != nil {
			return wd, err
		}
		if err := p.expect('='); err != nil {
			return wd, err
		}
		switch key {
		case "size":
			fs, err := p.tuple()
			if err != nil {
				return wd, err
			}
			if wd.Size, err = asVec2(fs, "size"); err != nil {
				return wd, p.errorf("%v", err)
			}
		case "offset":
			fs, err := p.tuple()
			if err != nil {
				return wd, err
			}
			if wd.Offset, err = asVec3(fs, "offset"); err != nil {
				return wd, p.errorf("%v", err)
			}
		default:
			return wd, p.errorf("unknown WalkableSurface property %q", key)
		}
	}
	return wd, nil
}

func (p *agparser) parseBlockerVolume(name string) (BlockerVolumeData, error) {
	bd := BlockerVolumeData{Name: name}
	for {
		p.skipWS()
		if p.eof() {
			return bd, p.errorf("unterminated BlockerVolume %q block", name)
		}
		if p.peek() == '}' {
			p.advance()
			break
		}
		key, err := p.ident()
		if err != nil {
			return bd, err
		}
		if err := p.expect('='); err != nil {
			return bd, err
		}
		switch key {
		case "size":
			fs, err := p.tuple()
			if err != nil {
				return bd, err
			}
			if bd.Size, err = asVec3(fs, "size"); err != nil {
				return bd, p.errorf("%v", err)
			}
		case "position":
			fs, err := p.tuple()
			if err != nil {
				return bd, err
			}
			if bd.Position, err = asVec3(fs, "position"); err != nil {
				return bd, p.errorf("%v", err)
			}
		default:
			return bd, p.errorf("unknown BlockerVolume property %q", key)
		}
	}
	return bd, nil
}

func (p *agparser) parseSpawnPoint(name string) (SpawnPointData, error) {
	sd := SpawnPointData{Name: name}
	for {
		p.skipWS()
		if p.eof() {
			return sd, p.errorf("unterminated SpawnPoint %q block", name)
		}
		if p.peek() == '}' {
			p.advance()
			break
		}
		key, err := p.ident()
		if err != nil {
			return sd, err
		}
		if err := p.expect('='); err != nil {
			return sd, err
		}
		switch key {
		case "character":
			v, err := p.str()
			if err != nil {
				return sd, err
			}
			sd.Character = v
		case "position":
			fs, err := p.tuple()
			if err != nil {
				return sd, err
			}
			if sd.Position, err = asVec3(fs, "position"); err != nil {
				return sd, p.errorf("%v", err)
			}
		default:
			return sd, p.errorf("unknown SpawnPoint property %q", key)
		}
	}
	return sd, nil
}

func (p *agparser) parseTriggerRegion(name string) (TriggerRegionData, error) {
	td := TriggerRegionData{Name: name}
	for {
		p.skipWS()
		if p.eof() {
			return td, p.errorf("unterminated TriggerRegion %q block", name)
		}
		if p.peek() == '}' {
			p.advance()
			break
		}
		key, err := p.ident()
		if err != nil {
			return td, err
		}
		if err := p.expect('='); err != nil {
			return td, err
		}
		switch key {
		case "size":
			fs, err := p.tuple()
			if err != nil {
				return td, err
			}
			if td.Size, err = asVec3(fs, "size"); err != nil {
				return td, p.errorf("%v", err)
			}
		case "position":
			fs, err := p.tuple()
			if err != nil {
				return td, err
			}
			if td.Position, err = asVec3(fs, "position"); err != nil {
				return td, p.errorf("%v", err)
			}
		default:
			return td, p.errorf("unknown TriggerRegion property %q", key)
		}
	}
	return td, nil
}

func (p *agparser) parseHotspot(name string) (HotspotData, error) {
	hd := HotspotData{Name: name}
	for {
		p.skipWS()
		if p.eof() {
			return hd, p.errorf("unterminated Hotspot %q block", name)
		}
		if p.peek() == '}' {
			p.advance()
			break
		}
		key, err := p.ident()
		if err != nil {
			return hd, err
		}
		if err := p.expect('='); err != nil {
			return hd, err
		}
		switch key {
		case "size":
			fs, err := p.tuple()
			if err != nil {
				return hd, err
			}
			if hd.Size, err = asVec3(fs, "size"); err != nil {
				return hd, p.errorf("%v", err)
			}
		case "position":
			fs, err := p.tuple()
			if err != nil {
				return hd, err
			}
			if hd.Position, err = asVec3(fs, "position"); err != nil {
				return hd, p.errorf("%v", err)
			}
		default:
			return hd, p.errorf("unknown Hotspot property %q", key)
		}
	}
	return hd, nil
}

// parseRoomDialogue reads the dialogue { ... } block fields.
func (p *agparser) parseRoomDialogue() (*RoomDialogue, error) {
	dlg := &RoomDialogue{}
	for {
		p.skipWS()
		if p.eof() {
			return nil, p.errorf("unterminated dialogue block — missing '}'")
		}
		if p.peek() == '}' {
			p.advance()
			return dlg, nil
		}
		key, err := p.ident()
		if err != nil {
			return nil, err
		}
		if err := p.expect('='); err != nil {
			return nil, err
		}
		v, err := p.str()
		if err != nil {
			return nil, err
		}
		switch key {
		case "on_enter":
			dlg.OnEnter = v
		case "on_enter_repeat":
			dlg.OnEnterRepeat = v
		default:
			return nil, p.errorf("unknown dialogue field %q", key)
		}
	}
}
