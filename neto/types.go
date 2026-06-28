package neto

type MapInstance int

const (
	MainMap MapInstance = 0
	DLC     MapInstance = 1
)

type MatchStatus int

const (
	NotRunning  MatchStatus = 0
	Starting    MatchStatus = 1
	Preparation MatchStatus = 2
	Running     MatchStatus = 3
	Finished    MatchStatus = 4
)

type EldenRingClasses int

const (
	Vagabond EldenRingClasses = 0
	Warrior  EldenRingClasses = 1
	Hero     EldenRingClasses = 2
	Bandit   EldenRingClasses = 3
	Astrologer EldenRingClasses = 4
	Prophet  EldenRingClasses = 5
	Samurai  EldenRingClasses = 6
	Prisoner EldenRingClasses = 7
	Confessor EldenRingClasses = 8
	Wretch   EldenRingClasses = 9
)

type UserInRoom struct {
	Guid    string `msgpack:"Guid"`
	IsAdmin bool   `msgpack:"IsAdmin"`
	Nick    string `msgpack:"Nick"`
	Team    int    `msgpack:"Team"`
}

type SquareCounter struct {
	Counter int `msgpack:"Counter"`
	Team    int `msgpack:"Team"`
}

type BingoBoardSquare struct {
	Text     string          `msgpack:"Text"`
	Tooltip  string          `msgpack:"Tooltip"`
	Team     []int           `msgpack:"Team"`
	Marked   bool            `msgpack:"Marked"`
	Counters []SquareCounter `msgpack:"Counters"`
}

type BingoGameSettings struct {
	BoardSize          int    `msgpack:"BoardSize"`
	Lockout            bool   `msgpack:"Lockout"`
	RandomClasses      bool   `msgpack:"RandomClasses"`
	ValidClasses       []int  `msgpack:"ValidClasses"`
	NumberOfClasses    int    `msgpack:"NumberOfClasses"`
	CategoryLimit      int    `msgpack:"CategoryLimit"`
	RandomSeed         int    `msgpack:"RandomSeed"`
	PreparationTime    int    `msgpack:"PreparationTime"`
	PointsPerBingoLine int    `msgpack:"PointsPerBingoLine"`
}

type TeamScore struct {
	Team  int    `msgpack:"Team"`
	Name  string `msgpack:"Name"`
	Score int    `msgpack:"Score"`
}

type BingoLine struct {
	Team       int    `msgpack:"Team"`
	Name       string `msgpack:"Name"`
	Type       int    `msgpack:"Type"`
	BingoIndex int    `msgpack:"BingoIndex"`
}
