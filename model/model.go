package model

import (
	"encoding/json"
)

type (
	Preferences struct {
		Champion map[string]Preference
	}

	Preference struct {
		Name      string    `json:"name"`
		OPGG      string    `json:"opgg"`
		ARAM      ARAM      `json:"aram"`
		Classic   Classic   `json:"classic"`
		OneForAll OneForAll `json:"oneforall"`
		URF       URF       `json:"urf"`
	}

	ARAM struct {
		X string `json:"x"`
		Y string `json:"y"`
	}

	Classic struct {
		X string `json:"x"`
		Y string `json:"y"`
	}

	OneForAll struct {
		X string `json:"x"`
		Y string `json:"y"`
	}

	URF struct {
		X string `json:"x"`
		Y string `json:"y"`
	}

	Runes struct {
		Primary   float64
		Secondary float64
		Runes     []float64
	}

	Champion struct {
		Type    string
		Format  string
		Version string
		Data    map[string]ChampionData
	}

	ChampionData struct {
		Version string
		Id      string
		Key     string
		Name    string
		Title   string
		Blurb   string

		Info struct {
			Attack     float64
			Defense    float64
			Magic      float64
			Difficulty float64
		}

		Image struct {
			Full   string
			Sprite string
			Group  string
			X      float64
			Y      float64
			W      float64
			H      float64
		}

		Tags    []string
		Partype string

		Stats struct {
			HP                   float64
			HPPerLevel           float64
			MP                   float64
			MPPerLevel           float64
			Movespeed            float64
			Armor                float64
			ArmorPerLevel        float64
			Spellblock           float64
			SpellblockPerLevel   float64
			AttackRange          float64
			HPRegen              float64
			HPRegenPerLevel      float64
			MPRegen              float64
			MPRegenPerLevel      float64
			Crit                 float64
			CritPerLevel         float64
			AttackDamage         float64
			AttackDamagePerLevel float64
			AttackSpeedPerLevel  float64
			AttackSpeed          float64
		}
	}

	// https://ddragon.leagueoflegends.com/realms/na.json
	Realms struct {
		CDN string
		CSS string
		DD  string
		L   string
		LG  string

		N struct {
			Champion    string
			Item        string
			Language    string
			Map         string
			Mastery     string
			ProfileIcon string
			Rune        string
			Sticker     string
			Summoner    string
		}

		ProfileIconMax float64
		Store          string
		V              string
	}

	// Method: 	UPDATE
	// URI: 	/lol-settings/v1/account/lol-collection-champions
	ChampionCollection struct {
		Data struct {
			Data struct {
				GroupingDropdownKey string
				LastVisitTime       float64
				SortingDropdownKey  string
				UnownedFilter       string
			}
			SchemaVersion float64
		}

		EventType string
		URI       string
	}

	// Method: 	GET
	// URI: 	/lol-login/v1/session
	Login struct {
		AccountId      float64
		Connected      bool
		Error          bool
		GasToken       string
		IdToken        string
		IsInLoginQueue bool
		IsNewPlayer    bool
		Puuid          string
		State          string
		SummonerId     float64
		UserAuthToken  string
		Username       string
	}

	// Method: 	GET
	// URI: 	/lol-perks/v1/inventory
	OwnedPages struct {
		Count float64 `json:"OwnedPageCount"`
	}

	// Method: 	GET
	// URI: 	/lol-perks/v1/pages
	Page struct {
		AutoModifiedSelections []float64
		Current                bool
		Id                     float64
		IsActive               bool
		IsDeletable            bool
		IsEditable             bool
		IsValid                bool
		LastModified           float64
		Name                   string
		Order                  float64
		PrimaryStyleId         float64
		SelectedPerkIds        []float64
		SubStyleId             float64
	}

	// Method: 	UPDATE
	// URI: 	/lol-matchmaking/v1/ready-check
	MatchFound struct {
		Data struct {
			DeclinerIds    []float64
			DodgeWarning   string
			PlayerResponse string
			State          string
			SuppressUx     bool
			Timer          float64
		}

		EventType string
		URI       string
	}

	// Method: 	UPDATE
	// URI: 	/lol-gameflow/v1/gameflow-phase
	Phase struct {
		Data      string
		EventType string
		URI       string
	}

	// Method: 	GET
	// URI: 	/lol-gameflow/v1/session
	Gameflow struct {
		GameClient struct {
			ObserverServerIp   string
			ObserverServerPort float64
			Running            bool
			ServerIp           string
			ServerPort         float64
			Visible            bool
		}

		GameData struct {
			GameId                   float64
			GameName                 string
			IsCustomGame             bool
			Password                 string
			PlayerChampionSelections interface{}

			Queue struct {
				AllowablePremadeSizes   interface{}
				AreFreeChampionsAllowed bool
				AssetMutator            string
				Category                string
				ChampionsRequiredToPlay float64
				Description             string
				DetailedDescription     string
				GameMode                string

				GameTypeConfig struct {
					AdvancedLearningQuests bool
					AllowTrades            bool
					BanMode                string
					BanTimerDuration       float64
					BattleBoost            bool
					CrossTeamChampionPool  bool
					DeathMatch             bool
					DoNotRemove            bool
					DuplicatePick          bool
					ExclusivePick          bool
					Id                     float64
					LearningQuests         bool
					MainPickTimerDuration  float64
					MaxAllowableBans       float64
					Name                   string
					OnboardCoopBeginner    bool
					PickMode               string
					PostPickTimerDuration  float64
					Reroll                 bool
					TeamChampionPool       bool
				}

				Id                                  float64
				IsRanked                            bool
				IsTeamBuilderManaged                bool
				IsTeamOnly                          bool
				LastToggledOffTime                  float64
				LastToggledOnTime                   float64
				MapId                               float64
				MaxLevel                            float64
				MaxSummonerLevelForFirstWinOfTheDay float64
				MaximumParticipantListSize          float64
				MinLevel                            float64
				MinimumParticipantListSize          float64
				Name                                string
				NumPlayersPerTeam                   float64
				QueueAvailability                   string

				QueueRewards struct {
					IsChampionPointsEnabled bool
					IsIpEnabled             bool
					IsXpEnabled             bool
					PartySizeIpRewards      []float64
				}

				RemovalFromGameAllowed      bool
				RemovalFromGameDelayMinutes float64
				ShortName                   string
				ShowPositionSelector        bool
				SpectatorEnabled            bool
				Type                        string
			}
			SpectatorsAllowed bool
			TeamOne           []interface{}
			TeamTwo           []interface{}
		}

		GameDodge struct {
			DodgeIds []float64
			Phase    string
			State    string
		}

		Map struct {
			Assets struct {
				ChampSelectBackgroundSound   string `json:"champ-select-background-sound"`
				ChampSelectFlyingBackground  string `json:"champ-select-flyout-background"`
				ChampSelectPlanningIntro     string `json:"champ-select-planning-intro"`
				GameSelectIconActive         string `json:"game-select-icon-active"`
				GameSelectIconActiveVideo    string `json:"game-select-icon-active-video"`
				GameSelectIconDefault        string `json:"game-select-icon-default"`
				GameSelectIconDisabled       string `json:"game-select-icon-disabled"`
				GameSelectIconHover          string `json:"game-select-icon-hover"`
				GameSelectIconIntroVideo     string `json:"game-select-icon-intro-video"`
				GameflowBackground           string `json:"gameflow-background"`
				GameflowBackgroundHoverSound string `json:"gameselect-button-hover-sound"`
				IconDefeat                   string `json:"icon-defeat"`
				IconDefeatVideo              string `json:"icon-defeat-video"`
				IconEmpty                    string `json:"icon-empty"`
				IconHover                    string `json:"icon-hover"`
				IconLeaver                   string `json:"icon-leaver"`
				IconVictory                  string `json:"icon-victory"`
				IconVictoryVideo             string `json:"icon-victory-video"`
				MapNorth                     string `json:"map-north"`
				MapSouth                     string `json:"map-south"`
				MusicInqueueLoopSound        string `json:"music-inqueue-loop-sound"`
				PartiesBackground            string `json:"parties-background"`
				PostgameAmbienceLoopSound    string `json:"postgame-ambience-loop-sound"`
				ReadyCheckBackground         string `json:"ready-check-background"`
				ReadyCheckBackgroundSound    string `json:"ready-check-background-sound"`
				SfxAmbiencePregameLoopSound  string `json:"sfx-ambience-pregame-loop-sound"`
				SocialIconLeaver             string `json:"social-icon-leaver"`
				SocialIconVictory            string `json:"social-icon-victory"`
			}
			CategorizedContentBundles           interface{}
			Description                         string
			GameMode                            string
			GameModeName                        string
			GameModeShortName                   string
			GameMutator                         string
			Id                                  float64
			IsRGM                               bool
			MapStringId                         string
			Name                                string
			PerPositionDisallowedSummonerSpells interface{}
			PerPositionRequiredSummonerSpells   interface{}
			PlatformId                          string
			PlatformName                        string
			Properties                          struct {
				SuppressRunesMasteriesPerks bool
			}
		}

		Phase string
	}

	// Method: 	GET
	// URI: 	/lol-champ-select/v1/session
	ChampionSelection struct {
		Data struct {
			Actions [][]struct {
				ActorCellId  float64
				ChampionId   float64
				Completed    bool
				Id           float64
				IsAllyAction bool
				IsInProgress bool
				PickTurn     float64
				ActionType   string `json:"type"`
			}

			AllowBattleBoost    bool
			AllowDuplicatePicks bool
			AllowLockedEvents   bool
			AllowRerolling      bool
			AllowSkinSelection  bool

			Bans struct {
				MyTeamBans    []float64
				NumBans       float64
				TheirTeamBans []float64
			}

			BenchChampionIds   []float64
			BenchEnabled       bool
			BoostableSkinCount float64

			ChatDetails struct {
				ChatRoomName     string
				ChatRoomPassword string
			}

			Counter float64

			EntitledFeatureState struct {
				AdditionalRerolls float64
				UnlockedSkinIds   []float64
			}

			GameId               float64
			HasSimultaneousBans  bool
			HasSimultaneousPicks bool
			IsCustomGame         bool
			IsSpectating         bool
			LocalPlayerCellId    float64
			LockedEventIndex     float64

			MyTeam []struct {
				AssignedPosition    string
				CellId              float64
				ChampionId          float64
				ChampionPickIntent  float64
				EntitledFeatureType string
				SelectedSkinId      float64
				Spell1Id            float64
				Spell2Id            float64
				SummonerId          float64
				Team                float64
				WardSkinId          float64
			}

			RecoveryCounter    float64
			RerollsRemaining   float64
			SkipChampionSelect bool

			TheirTeam []struct {
				AssignedPosition    string
				CellId              float64
				ChampionId          float64
				ChampionPickIntent  float64
				EntitledFeatureType string
				SelectedSkinId      float64
				Spell1Id            float64
				Spell2Id            float64
				SummonerId          float64
				Team                float64
				WardSkinId          float64
			}

			Timer struct {
				AdjustedTimeLeftInPhase float64
				InternalNowInEpochMs    float64
				IsInfinite              bool
				Phase                   string
				TotalTimeInPhase        float64
			}

			Trades struct {
				CellId float64
				Id     float64
				State  string
			}
		}

		EventType string
		URI       string
	}
)

func (preferences Preferences) MarshalJSON() ([]byte, error) {
	champion, err := json.Marshal(preferences.Champion)
	return []byte(string(champion)), err
}
