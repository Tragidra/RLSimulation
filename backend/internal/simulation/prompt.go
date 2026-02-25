package simulation

import (
	"fmt"
	"strings"

	"simarena/internal/llm"
	"simarena/internal/models"
)

// BuildAgentRoundMessages constructs chat messages for a specific agent in a given round.
func BuildAgentRoundMessages(sim *models.Simulation, agent models.Agent, round int) []llm.ChatMessage {
	interactive := sim.IsInteractive()
	ru := sim.Language == "ru"

	var sys strings.Builder

	// Agent identity
	if ru {
		sys.WriteString("Ты агент в симуляции.\n")
		sys.WriteString(fmt.Sprintf("Твоё имя: %s\n", agent.Name))
		if agent.Role != "" {
			sys.WriteString(fmt.Sprintf("Твоя роль: %s\n", agent.Role))
		}
		sys.WriteString(fmt.Sprintf("\nСценарий: %s\n", sim.Description))
		sys.WriteString(fmt.Sprintf("Предусловия: %s\n", sim.Preconditions))
	} else {
		sys.WriteString("You are an agent in a simulation.\n")
		sys.WriteString(fmt.Sprintf("Your name: %s\n", agent.Name))
		if agent.Role != "" {
			sys.WriteString(fmt.Sprintf("Your role: %s\n", agent.Role))
		}
		sys.WriteString(fmt.Sprintf("\nScenario: %s\n", sim.Description))
		sys.WriteString(fmt.Sprintf("Preconditions: %s\n", sim.Preconditions))
	}

	// Other participants (interactive mode only)
	if interactive {
		if ru {
			sys.WriteString("\nДругие участники:\n")
		} else {
			sys.WriteString("\nOther participants:\n")
		}
		for _, other := range sim.Agents {
			if other.ID == agent.ID {
				continue
			}
			role := other.Role
			if role == "" {
				if ru {
					role = "участник"
				} else {
					role = "participant"
				}
			}
			sys.WriteString(fmt.Sprintf("- %s: %s\n", other.Name, role))
		}
	}

	// Round info
	if ru {
		sys.WriteString(fmt.Sprintf("\nВ этой симуляции %d раундов. Текущий раунд: %d/%d.\n", sim.Rounds, round, sim.Rounds))
	} else {
		sys.WriteString(fmt.Sprintf("\nThis simulation has %d rounds. Current round: %d/%d.\n", sim.Rounds, round, sim.Rounds))
	}

	// History with context window strategy
	historySteps := buildAgentContext(sim, agent.ID, round, interactive)
	truncated := false
	if sim.Rounds > 5 && round > 3 {
		truncated = true
	}

	if len(historySteps) > 0 {
		if ru {
			sys.WriteString("\n=== История ===\n")
		} else {
			sys.WriteString("\n=== History ===\n")
		}

		currentHistRound := 0
		for _, step := range historySteps {
			if step.Round != currentHistRound {
				currentHistRound = step.Round
				sys.WriteString(fmt.Sprintf("--- Раунд %d ---\n", step.Round))
			}
			sys.WriteString(fmt.Sprintf("[%s]: %s\n", step.AgentName, step.Content))
		}

		if truncated {
			if ru {
				sys.WriteString("(Для эффективности показаны только последние 2 раунда. Более ранняя история опущена.)\n")
			} else {
				sys.WriteString("(Only last 2 rounds shown for context efficiency)\n")
			}
		}

		if ru {
			sys.WriteString("=== Конец истории ===\n")
		} else {
			sys.WriteString("=== End History ===\n")
		}
	}

	// User message
	var user string
	if interactive {
		if ru {
			sys.WriteString(fmt.Sprintf("\nВыполни раунд %d. Учитывай действия других агентов и развивающуюся ситуацию. Опиши свой анализ, решения и действия. Оставайся в роли %s", round, agent.Name))
			if agent.Role != "" {
				sys.WriteString(fmt.Sprintf(" (%s)", agent.Role))
			}
			sys.WriteString(".\n")
			user = fmt.Sprintf("Выполни раунд %d. Отвечай на русском языке.", round)
		} else {
			user = fmt.Sprintf("Execute round %d. Consider other agents' actions and the evolving situation. Describe your analysis, decisions, and actions. Stay in character as %s", round, agent.Name)
			if agent.Role != "" {
				user += fmt.Sprintf(" (%s)", agent.Role)
			}
			user += ".\nRespond in English."
		}
	} else {
		if ru {
			user = fmt.Sprintf("Выполни раунд %d. Опиши свой анализ, решения и действия. Будь подробным и стратегическим. Отвечай на русском языке.", round)
		} else {
			user = fmt.Sprintf("Execute round %d. Describe your analysis, decisions, and actions for this round. Be detailed and strategic.\nRespond in English.", round)
		}
	}

	return []llm.ChatMessage{
		{Role: "system", Content: sys.String()},
		{Role: "user", Content: user},
	}
}

// buildAgentContext returns the history steps an agent should see, applying the context window strategy.
func buildAgentContext(sim *models.Simulation, agentID string, currentRound int, interactive bool) []models.Step {
	minRound := 1
	if sim.Rounds > 5 && currentRound > 3 {
		minRound = currentRound - 2
	}

	var result []models.Step
	for _, step := range sim.Steps {
		if step.Round >= currentRound {
			// Include steps from current round by agents that already acted this round
			if step.Round == currentRound {
				if interactive {
					result = append(result, step)
				}
			}
			continue
		}
		if step.Round < minRound {
			continue
		}
		if interactive {
			result = append(result, step)
		} else {
			// Independent mode: only own steps
			if step.AgentID == agentID {
				result = append(result, step)
			}
		}
	}
	return result
}

// BuildSummaryMessages constructs the chat messages to generate a final summary.
func BuildSummaryMessages(sim *models.Simulation) []llm.ChatMessage {
	ru := sim.Language == "ru"

	var sys strings.Builder
	if ru {
		sys.WriteString("Ты агент симуляции. Ты завершил симуляцию со следующим сценарием:\n\n")
		sys.WriteString(fmt.Sprintf("Сценарий: %s\n", sim.Description))
		sys.WriteString(fmt.Sprintf("Предусловия: %s\n\n", sim.Preconditions))
		sys.WriteString("Вот лог всех раундов:\n")
	} else {
		sys.WriteString("You are a simulation agent. You completed a simulation with the following scenario:\n\n")
		sys.WriteString(fmt.Sprintf("Scenario: %s\n", sim.Description))
		sys.WriteString(fmt.Sprintf("Preconditions: %s\n\n", sim.Preconditions))
		sys.WriteString("Here is a log of all rounds:\n")
	}

	currentRound := 0
	for _, step := range sim.Steps {
		if step.Round != currentRound {
			currentRound = step.Round
			sys.WriteString(fmt.Sprintf("\n--- Round %d ---\n", step.Round))
		}
		sys.WriteString(fmt.Sprintf("[%s]: %s\n", step.AgentName, step.Content))
	}

	var user string
	if ru {
		user = "Предоставь итоговое резюме симуляции. Каков был общий результат? Какие были ключевые решения и поворотные моменты? Каково итоговое состояние? Отвечай на русском языке."
	} else {
		user = "Provide a final summary of the simulation. What was the overall outcome? What were the key decisions and turning points? What was the final state? Respond in English."
	}

	return []llm.ChatMessage{
		{Role: "system", Content: sys.String()},
		{Role: "user", Content: user},
	}
}
