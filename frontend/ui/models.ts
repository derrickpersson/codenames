export interface IPlayer {
    player_name: string;
    team: Team;
}

export type Team = "blue" | "red";

export interface TeamPoint {
	team: Team;
	points: number;
}

export interface IGame {
    created_at: string;  // Date time string
    current_player: number; // Index in routing order
    current_word: string;
    enforce_timer?: boolean; // Phasing out
    id: string;
    layout?: null; // Phasing out
    perm_index: number;
    revealed: boolean[]; 
    round: number;
    round_started_at: string; // Date time string
    routing_order: IPlayer[];
    seed: number; // Random int
    stage: number; // 0, 1, 2, 3 => Setup, Explain, OneWord, Gestures
    starting_team: Team;
    state_id: string;
    team_players: IPlayer[];
    timer_duration_ms: number;
    updated_at: string; // Date time string
    word_set: string[];
    words: string[];
    team_points: TeamPoint[];
}