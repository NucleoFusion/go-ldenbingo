package neto

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/vmihailenco/msgpack/v5"
)

type PacketType byte

const (
	ServerRegisterAccepted PacketType = 0
	ServerRegisterDenied   PacketType = 1
	ServerClientDropped    PacketType = 2
	ServerShutdown         PacketType = 3
	ClientRegister         PacketType = 4
	ClientDisconnect       PacketType = 5
	ObjectData             PacketType = 6
	KeepAlive              PacketType = 7
)

type Packet struct {
	Type    PacketType
	Objects []interface{}
}

// System Neto packet structures (mirroring C# Client/Server register structures)

type ServerRegisterAcceptedPayload struct {
	Message    string `msgpack:"Message"`
	ClientGuid string `msgpack:"ClientGuid"`
}

type ServerRegisterDeniedPayload struct {
	Message string `msgpack:"Message"`
}

type ClientRegisterPayload struct {
	Message       string `msgpack:"Message"`
	Version       string `msgpack:"Version"`
	IdentityToken string `msgpack:"IdentityToken"`
}

type ServerKickedPayload struct {
	Reason string `msgpack:"Reason"`
}

type KeepAlivePayload struct{}

// EldenBingo Packet structures

type ServerRoomNameSuggestion struct {
	RoomName string `msgpack:"RoomName"`
}

type ServerCreateRoomDenied struct {
	Reason string `msgpack:"Reason"`
}

type ServerJoinRoomAccepted struct {
	RoomName    string       `msgpack:"RoomName"`
	Users       []UserInRoom `msgpack:"Users"`
	MatchStatus MatchStatus  `msgpack:"MatchStatus"`
	Paused      bool         `msgpack:"Paused"`
	Timer       int          `msgpack:"Timer"`
}

type ServerJoinRoomDenied struct {
	Reason string `msgpack:"Reason"`
}

type ServerUserJoinedRoom struct {
	User UserInRoom `msgpack:"User"`
}

type ServerUserLeftRoom struct {
	User UserInRoom `msgpack:"User"`
}

type ServerUserCoordinates struct {
	UserGuid      string      `msgpack:"UserGuid"`
	X             float32     `msgpack:"X"`
	Y             float32     `msgpack:"Y"`
	Angle         float32     `msgpack:"Angle"`
	IsUnderground bool        `msgpack:"IsUnderground"`
	MapInstance   MapInstance `msgpack:"MapInstance"`
}

type ServerAdminStatusMessage struct {
	Message string `msgpack:"Message"`
	Color   int    `msgpack:"Color"`
}

type ServerUserChat struct {
	UserGuid string `msgpack:"UserGuid"`
	Message  string `msgpack:"Message"`
}

type ServerMatchStatusUpdate struct {
	MatchStatus MatchStatus `msgpack:"MatchStatus"`
	Paused      bool        `msgpack:"Paused"`
	Timer       int         `msgpack:"Timer"`
}

type ServerEntireBingoBoardUpdate struct {
	Size             int                `msgpack:"Size"`
	Lockout          bool               `msgpack:"Lockout"`
	Squares          []BingoBoardSquare `msgpack:"Squares"`
	AvailableClasses []int              `msgpack:"AvailableClasses"`
}

type ServerScoreboardUpdate struct {
	Scoreboard []TeamScore `msgpack:"Scoreboard"`
}

type ServerBingoAchievedUpdate struct {
	Bingo BingoLine `msgpack:"Bingo"`
}

type ServerSquareUpdate struct {
	Square BingoBoardSquare `msgpack:"Square"`
	Index  int              `msgpack:"Index"`
}

type ServerUserChecked struct {
	UserGuid     string `msgpack:"UserGuid"`
	Index        int    `msgpack:"Index"`
	Team         int    `msgpack:"Team"`
	TeamsChecked []int  `msgpack:"TeamsChecked"`
}

type ServerCurrentGameSettings struct {
	GameSettings BingoGameSettings `msgpack:"GameSettings"`
}

type ServerTeamNameChanged struct {
	UserGuid      string `msgpack:"UserGuid"`
	Team          int    `msgpack:"Team"`
	TeamColorName string `msgpack:"TeamColorName"`
	Name          string `msgpack:"Name"`
}

type ServerBroadcastMessage struct {
	Message string `msgpack:"Message"`
}

type ServerUserChangedTeam struct {
	UserGuid      string       `msgpack:"UserGuid"`
	Team          int          `msgpack:"Team"`
	TeamColorName string       `msgpack:"TeamColorName"`
	Users         []UserInRoom `msgpack:"Users"`
}

type ServerUserBannedFromRoom struct {
	User   UserInRoom `msgpack:"User"`
	Banner UserInRoom `msgpack:"Banner"`
}

type ServerPromoteToAdmin struct {
	User     UserInRoom `msgpack:"User"`
	Promoter UserInRoom `msgpack:"Promoter"`
}

type ClientRequestRoomName struct{}

type ClientRequestCreateRoom struct {
	RoomName  string            `msgpack:"RoomName"`
	AdminPass string            `msgpack:"AdminPass"`
	Nick      string            `msgpack:"Nick"`
	Team      int               `msgpack:"Team"`
	Settings  BingoGameSettings `msgpack:"Settings"`
}

type ClientRequestJoinRoom struct {
	RoomName  string `msgpack:"RoomName"`
	AdminPass string `msgpack:"AdminPass"`
	Nick      string `msgpack:"Nick"`
	Team      int    `msgpack:"Team"`
}

type ClientRequestLeaveRoom struct{}

type ClientCoordinates struct {
	X             float32     `msgpack:"X"`
	Y             float32     `msgpack:"Y"`
	Angle         float32     `msgpack:"Angle"`
	IsUnderground bool        `msgpack:"IsUnderground"`
	MapInstance   MapInstance `msgpack:"MapInstance"`
}

type ClientChat struct {
	Message string `msgpack:"Message"`
}

type ClientBingoJson struct {
	Json string `msgpack:"Json"`
}

type ClientRandomizeBoard struct{}

type ClientChangeMatchStatus struct {
	MatchStatus MatchStatus `msgpack:"MatchStatus"`
}

type ClientTogglePause struct{}

type ClientTryCheck struct {
	Index   int    `msgpack:"Index"`
	ForUser string `msgpack:"ForUser"`
}

type ClientTryMark struct {
	Index int `msgpack:"Index"`
}

type ClientTrySetCounter struct {
	Index   int    `msgpack:"Index"`
	Change  int    `msgpack:"Change"`
	ForUser string `msgpack:"ForUser"`
}

type ClientSetGameSettings struct {
	GameSettings BingoGameSettings `msgpack:"GameSettings"`
}

type ClientRequestCurrentGameSettings struct{}

type ClientSetTeamName struct {
	Team int    `msgpack:"Team"`
	Name string `msgpack:"Name"`
}

type ClientRequestTeamChange struct {
	Team int `msgpack:"Team"`
}

type ClientBanUserFromRoom struct {
	BannedUser string `msgpack:"BannedUser"`
}

type ClientPromoteToAdmin struct {
	PromotedUser string `msgpack:"PromotedUser"`
}

// Registry maps C# FullName string to Go reflect.Type
var typeRegistry = make(map[string]reflect.Type)
var nameRegistry = make(map[reflect.Type]string)

func RegisterType(name string, val interface{}) {
	t := reflect.TypeOf(val)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	typeRegistry[name] = t
	nameRegistry[t] = name
}

func init() {
	RegisterType("Neto.Shared.ServerRegisterAccepted", ServerRegisterAcceptedPayload{})
	RegisterType("Neto.Shared.ServerRegisterDenied", ServerRegisterDeniedPayload{})
	RegisterType("Neto.Shared.ClientRegister", ClientRegisterPayload{})
	RegisterType("Neto.Shared.ServerKicked", ServerKickedPayload{})
	RegisterType("Neto.Shared.KeepAlive", KeepAlivePayload{})

	RegisterType("ServerRoomNameSuggestion", ServerRoomNameSuggestion{})
	RegisterType("ServerCreateRoomDenied", ServerCreateRoomDenied{})
	RegisterType("ServerJoinRoomAccepted", ServerJoinRoomAccepted{})
	RegisterType("ServerJoinRoomDenied", ServerJoinRoomDenied{})
	RegisterType("ServerUserJoinedRoom", ServerUserJoinedRoom{})
	RegisterType("ServerUserLeftRoom", ServerUserLeftRoom{})
	RegisterType("ServerUserCoordinates", ServerUserCoordinates{})
	RegisterType("ServerAdminStatusMessage", ServerAdminStatusMessage{})
	RegisterType("ServerUserChat", ServerUserChat{})
	RegisterType("ServerMatchStatusUpdate", ServerMatchStatusUpdate{})
	RegisterType("ServerEntireBingoBoardUpdate", ServerEntireBingoBoardUpdate{})
	RegisterType("ServerScoreboardUpdate", ServerScoreboardUpdate{})
	RegisterType("ServerBingoAchievedUpdate", ServerBingoAchievedUpdate{})
	RegisterType("ServerSquareUpdate", ServerSquareUpdate{})
	RegisterType("ServerUserChecked", ServerUserChecked{})
	RegisterType("ServerCurrentGameSettings", ServerCurrentGameSettings{})
	RegisterType("ServerTeamNameChanged", ServerTeamNameChanged{})
	RegisterType("ServerBroadcastMessage", ServerBroadcastMessage{})
	RegisterType("ServerUserChangedTeam", ServerUserChangedTeam{})
	RegisterType("ServerUserBannedFromRoom", ServerUserBannedFromRoom{})
	RegisterType("ServerPromoteToAdmin", ServerPromoteToAdmin{})

	RegisterType("ClientRequestRoomName", ClientRequestRoomName{})
	RegisterType("ClientRequestCreateRoom", ClientRequestCreateRoom{})
	RegisterType("ClientRequestJoinRoom", ClientRequestJoinRoom{})
	RegisterType("ClientRequestLeaveRoom", ClientRequestLeaveRoom{})
	RegisterType("ClientCoordinates", ClientCoordinates{})
	RegisterType("ClientChat", ClientChat{})
	RegisterType("ClientBingoJson", ClientBingoJson{})
	RegisterType("ClientRandomizeBoard", ClientRandomizeBoard{})
	RegisterType("ClientChangeMatchStatus", ClientChangeMatchStatus{})
	RegisterType("ClientTogglePause", ClientTogglePause{})
	RegisterType("ClientTryCheck", ClientTryCheck{})
	RegisterType("ClientTryMark", ClientTryMark{})
	RegisterType("ClientTrySetCounter", ClientTrySetCounter{})
	RegisterType("ClientSetGameSettings", ClientSetGameSettings{})
	RegisterType("ClientRequestCurrentGameSettings", ClientRequestCurrentGameSettings{})
	RegisterType("ClientSetTeamName", ClientSetTeamName{})
	RegisterType("ClientRequestTeamChange", ClientRequestTeamChange{})
	RegisterType("ClientBanUserFromRoom", ClientBanUserFromRoom{})
	RegisterType("ClientPromoteToAdmin", ClientPromoteToAdmin{})
}

// Custom marshaling for C# MessagePack array structure: [Type, ObjectsCount, TypeName1, ObjectData1, ...]
var _ msgpack.Marshaler = (*Packet)(nil)
var _ msgpack.Unmarshaler = (*Packet)(nil)

func (p *Packet) MarshalMsgpack() ([]byte, error) {
	arrLen := 2 + len(p.Objects)*2
	slice := make([]interface{}, 0, arrLen)
	slice = append(slice, byte(p.Type))
	slice = append(slice, int32(len(p.Objects)))

	for _, obj := range p.Objects {
		t := reflect.TypeOf(obj)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		name, ok := nameRegistry[t]
		if !ok {
			return nil, fmt.Errorf("type %v is not registered in nameRegistry", t)
		}
		slice = append(slice, name)
		slice = append(slice, obj)
	}

	return msgpack.Marshal(slice)
}

func (p *Packet) UnmarshalMsgpack(b []byte) error {
	dec := msgpack.NewDecoder(bytes.NewReader(b))
	arrLen, err := dec.DecodeArrayLen()
	if err != nil {
		return err
	}
	if arrLen < 2 {
		return fmt.Errorf("invalid packet array length: %d", arrLen)
	}
	ptype, err := dec.DecodeUint8()
	if err != nil {
		return err
	}
	p.Type = PacketType(ptype)
	objCount, err := dec.DecodeInt32()
	if err != nil {
		return err
	}
	p.Objects = make([]interface{}, 0, objCount)
	for i := 0; i < int(objCount); i++ {
		typeName, err := dec.DecodeString()
		if err != nil {
			return err
		}
		t, ok := typeRegistry[typeName]
		if !ok {
			return fmt.Errorf("unregistered incoming type name: %s", typeName)
		}
		valPtr := reflect.New(t).Interface()
		if err := dec.Decode(valPtr); err != nil {
			return fmt.Errorf("failed to decode object of type %s: %w", typeName, err)
		}
		p.Objects = append(p.Objects, reflect.ValueOf(valPtr).Elem().Interface())
	}
	return nil
}
