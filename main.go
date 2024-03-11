package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/basedalex/deck"
)

// Hand is a slice of cards, represents players' hands
type Hand []deck.Card 

func (h Hand) String() string {
	strs := make([]string, len(h))

	for i := range h {
		strs[i] = h[i].String()
	}

	return strings.Join(strs, ", ")
}

func (h Hand) DealerString() string {
	return h[0].String() + ", **HIDDEN**"
}

func (h Hand) Score() int {
	minScore := h.MinScore()

	if minScore > 11 {
		return minScore
	}
	
	for _, c := range h {
		if c.Rank == deck.Ace {
			// ace is currently worth 1
			return minScore + 10
		}
	}
	return minScore
}

func (h Hand) MinScore() int {
	score := 0
	for _, c := range h {
		score += min(int(c.Rank), 10)
	}
	return score
}

func min(a, b int) int {
	if a < b {
		return a 
	} 
	return b
} 

func Shuffle(gs GameState) GameState {
	ret := clone(gs)
	ret.Deck = deck.New(deck.Deck(3), deck.Shuffle)
	return ret
}

func Deal(gs GameState) GameState {
	ret := clone(gs)
	ret.Player = make(Hand, 0, 5)
	ret.Dealer = make(Hand, 0, 5)
	var card deck.Card
	for i := 0; i < 2; i++ {
		card, ret.Deck = draw(ret.Deck)
		ret.Player = append(ret.Player, card)
		card, ret.Deck = draw(ret.Deck)
		ret.Dealer = append(ret.Dealer, card)
	}
	ret.State = StatePlayerTurn
	return ret
}

func Stand(gs GameState) GameState {
	ret := clone(gs)
	ret.State++
	return ret
}

func Hit(gs GameState) GameState {
	ret := clone(gs)
	hand := ret.CurrentPlayer()
	var card deck.Card
	card, ret.Deck = draw(ret.Deck)
	*hand = append(*hand, card)
	if hand.Score() > 21 {
		return Stand(ret)
	}
	if hand.Score() == 21 {
		return EndHand(ret)
	}
	return ret
}

func PlaceBets(gs GameState, input string) (GameState, error) {
	ret := clone(gs)
	bet, err := strconv.Atoi(input)
	if err != nil {
		return ret, errors.New("couldn't parse the input, please select a number")
	}
	if bet > ret.PlayerChips {
		return ret, errors.New("the bet is higher than your limit")
	}
	ret.PlayerBet = bet
	fmt.Println("Your current bet is", bet)
	return ret, nil
}

func EndHand(gs GameState) GameState {
	ret := clone(gs)
	bet := ret.PlayerBet
	fmt.Println("your bet is", bet)
	pScore, dScore := ret.Player.Score(), ret.Dealer.Score()
	fmt.Println("==FINAL HANDS==")
	fmt.Println("Player:", ret.Player, "\nScore:", pScore)
	fmt.Println("Dealer:", ret.Dealer, "\nScore:", dScore)
	switch {
	case pScore > 21:
		ret.PlayerChips = ret.PlayerChips - bet
		fmt.Printf("You busted, current chips %d\n", ret.PlayerChips)
	case dScore > 21:
		ret.PlayerChips = ret.PlayerChips + bet
		fmt.Printf("Dealer busted. current chips %d\n", ret.PlayerChips)
	case pScore > dScore:
		ret.PlayerChips = ret.PlayerChips + bet
		fmt.Printf("You win! Current chips %d\n", ret.PlayerChips)
	case dScore > pScore: 
		ret.PlayerChips = ret.PlayerChips - bet
		fmt.Printf("You lose Current chips %d\n", ret.PlayerChips)
	case dScore == pScore:
		fmt.Println("Draw")
	}
	fmt.Print()
	ret.Player = nil
	ret.Dealer = nil
	ret.PlayerBet = 0
	return ret
}

func main() {
	var gs GameState
	gs.PlayerChips = 200
	gs = Shuffle(gs)
	
	for i := 0; i < 10; i++ {
		fmt.Println("Place your bet! Current Balance is", gs.PlayerChips)
		var err error
		var bet string 
		fmt.Scanf("%s\n", &bet)

		gs, err = PlaceBets(gs, bet)
		for err != nil {
			fmt.Println(err)
			fmt.Scanf("%s\n", &bet)
			gs, err = PlaceBets(gs, bet)
		}
		
		gs = Deal(gs)
		var input string 

		for gs.State == StatePlayerTurn { 
			fmt.Println("Player:", gs.Player)
			fmt.Println("Dealer:", gs.Dealer.DealerString())
			fmt.Println("What will you do? (h)it, (s)tand")
			fmt.Scanf("%s\n", &input)
			switch input {
			case "h":
				gs = Hit(gs)
			case "s":
				gs = Stand(gs)
			default:
				fmt.Printf("unknown command %s, press (h) to (h)it or (s) to (s)tand", input)
			}
		}

		for gs.State == StateDealerTurn {
			if gs.Dealer.Score() <= 16 || (gs.Dealer.Score() == 17 && gs.Dealer.MinScore() != 17) {
				gs = Hit(gs)
			} else {
				gs = Stand(gs)
			}
		}
		
		gs = EndHand(gs)
	}
}

func draw(cards []deck.Card) (deck.Card, []deck.Card) {
	return cards[0], cards[1:]
}

type State int8
const (
	StatePlayerTurn State = iota
	StateDealerTurn
	StateHandOver
)

type GameState struct {
	Deck []deck.Card
	State State 
	Player Hand
	Dealer Hand
	PlayerChips int
	PlayerBet int
}

func (gs *GameState) CurrentPlayer() *Hand {
	switch gs.State {
	case StatePlayerTurn:
		return &gs.Player
	case StateDealerTurn:
		return &gs.Dealer
	default:
		panic("it is not any player's turn")
	}
}

func clone(gs GameState) GameState {
	ret := GameState {
		Deck: make([]deck.Card, len(gs.Deck)),
		State: gs.State,
		Player: make(Hand, len(gs.Player)),
		Dealer: make(Hand, len(gs.Dealer)),
		PlayerBet: gs.PlayerBet,
		PlayerChips: gs.PlayerChips,
	}
	copy(ret.Deck, gs.Deck)
	copy(ret.Player, gs.Player)
	copy(ret.Dealer, gs.Dealer)
	return ret
}