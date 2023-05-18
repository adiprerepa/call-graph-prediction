package main

type Match struct {
	Headers map[string]string `json:"headers"`
}

type Rule struct {
	OverrideHost string `json:"overrideHost"`
	Match        Match  `json:"match"`
}

type RouteMatchRules struct {
	Rule []Rule `json:"rules"`
}
