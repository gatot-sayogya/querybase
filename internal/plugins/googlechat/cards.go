package googlechat

import (
	"fmt"
	"time"

	"github.com/yourorg/querybase/internal/models"
	"google.golang.org/api/chat/v1"
)

// BuildSubmittedCard creates the initial approval request card
// Matches GPOS DBHUB format: Action, Query ID, Submitter, Query Type, Timestamp, Query
// With interactive buttons: Approve, Reject, View Query
func BuildSubmittedCard(approval *models.ApprovalRequest, appURL string) *chat.CardWithId {
	approvalURL := fmt.Sprintf("%s/dashboard/approvals/%s", appURL, approval.ID.String())
	shortID := shortUUID(approval.ID.String())

	return &chat.CardWithId{
		CardId: "approval_submitted_" + approval.ID.String(),
		Card: &chat.GoogleAppsCardV1Card{
			Header: &chat.GoogleAppsCardV1CardHeader{
				Title:    fmt.Sprintf("%s %s – %s", operationIcon(approval.OperationType), approval.OperationType, approval.DataSource.Name),
				Subtitle: "QueryBase",
				ImageUrl: "", // Can be set to logo URL
			},
			Sections: []*chat.GoogleAppsCardV1Section{
				{
					Widgets: []*chat.GoogleAppsCardV1Widget{
						decoratedText("Action", "Query Submitted"),
						decoratedText("Query ID", "#"+shortID),
						decoratedText("Submitter", approval.RequestedByUser.FullName),
						decoratedText("Query Type", string(approval.OperationType)),
						decoratedText("Timestamp", formatTimestamp(approval.CreatedAt)),
					},
				},
				{
					Header: "Query",
					Widgets: []*chat.GoogleAppsCardV1Widget{
						{
							TextParagraph: &chat.GoogleAppsCardV1TextParagraph{
								Text: fmt.Sprintf("<font color=\"#666666\"><code>%s</code></font>", truncateSQL(approval.QueryText, 500)),
							},
						},
					},
				},
				{
					Widgets: []*chat.GoogleAppsCardV1Widget{
						{
							ButtonList: &chat.GoogleAppsCardV1ButtonList{
								Buttons: []*chat.GoogleAppsCardV1Button{
									actionButton("✅ Approve", "action_approve", approval.ID.String(), "#1B8A2D"),
									actionButton("❌ Reject", "action_reject", approval.ID.String(), "#D32F2F"),
									linkButton("👁 View Query", approvalURL),
								},
							},
						},
					},
				},
			},
		},
	}
}

// BuildApprovedCard creates the "Query Approved" card for thread reply
func BuildApprovedCard(approval *models.ApprovalRequest, reviewer *models.User) *chat.CardWithId {
	shortID := shortUUID(approval.ID.String())

	return &chat.CardWithId{
		CardId: "approval_approved_" + approval.ID.String(),
		Card: &chat.GoogleAppsCardV1Card{
			Header: &chat.GoogleAppsCardV1CardHeader{
				Title:    fmt.Sprintf("%s %s – %s", operationIcon(approval.OperationType), approval.OperationType, approval.DataSource.Name),
				Subtitle: "QueryBase",
			},
			Sections: []*chat.GoogleAppsCardV1Section{
				{
					Widgets: []*chat.GoogleAppsCardV1Widget{
						decoratedText("Action", "Query Approved"),
						decoratedText("Query ID", "#"+shortID),
						decoratedText("Submitter", approval.RequestedByUser.FullName),
						decoratedText("Approver", reviewer.FullName),
						decoratedText("Query Type", string(approval.OperationType)),
						decoratedText("Timestamp", formatTimestamp(time.Now())),
					},
				},
				{
					Header: "Query",
					Widgets: []*chat.GoogleAppsCardV1Widget{
						{
							TextParagraph: &chat.GoogleAppsCardV1TextParagraph{
								Text: fmt.Sprintf("<font color=\"#666666\"><code>%s</code></font>", truncateSQL(approval.QueryText, 500)),
							},
						},
					},
				},
			},
		},
	}
}

// BuildRejectedCard creates the "Query Rejected" card for thread reply
func BuildRejectedCard(approval *models.ApprovalRequest, reviewer *models.User, reason string) *chat.CardWithId {
	shortID := shortUUID(approval.ID.String())

	widgets := []*chat.GoogleAppsCardV1Widget{
		decoratedText("Action", "Query Rejected"),
		decoratedText("Query ID", "#"+shortID),
		decoratedText("Submitter", approval.RequestedByUser.FullName),
		decoratedText("Reviewer", reviewer.FullName),
		decoratedText("Query Type", string(approval.OperationType)),
		decoratedText("Timestamp", formatTimestamp(time.Now())),
	}

	if reason != "" {
		widgets = append(widgets, decoratedText("Reason", reason))
	}

	return &chat.CardWithId{
		CardId: "approval_rejected_" + approval.ID.String(),
		Card: &chat.GoogleAppsCardV1Card{
			Header: &chat.GoogleAppsCardV1CardHeader{
				Title:    fmt.Sprintf("❌ %s – %s", approval.OperationType, approval.DataSource.Name),
				Subtitle: "QueryBase",
			},
			Sections: []*chat.GoogleAppsCardV1Section{
				{
					Widgets: widgets,
				},
			},
		},
	}
}

// BuildPreviewCard creates the transaction preview card with Commit/Rollback buttons
func BuildPreviewCard(approval *models.ApprovalRequest, txn *models.QueryTransaction) *chat.CardWithId {
	return &chat.CardWithId{
		CardId: "transaction_preview_" + txn.ID.String(),
		Card: &chat.GoogleAppsCardV1Card{
			Header: &chat.GoogleAppsCardV1CardHeader{
				Title:    "Transaction Preview",
				Subtitle: "QueryBase",
			},
			Sections: []*chat.GoogleAppsCardV1Section{
				{
					Widgets: []*chat.GoogleAppsCardV1Widget{
						decoratedText("Status", "Transaction Active"),
						decoratedText("Affected Rows", fmt.Sprintf("%d", txn.AffectedRows)),
						decoratedText("Started At", formatTimestamp(txn.StartedAt)),
					},
				},
				buildPreviewDataSection(txn),
				{
					Widgets: []*chat.GoogleAppsCardV1Widget{
						{
							ButtonList: &chat.GoogleAppsCardV1ButtonList{
								Buttons: []*chat.GoogleAppsCardV1Button{
									actionButton("✅ Commit", "action_commit", txn.ID.String(), "#1B8A2D"),
									actionButton("🔄 Rollback", "action_rollback", txn.ID.String(), "#FF6D00"),
								},
							},
						},
					},
				},
			},
		},
	}
}

// BuildCommittedCard creates the "Transaction Committed" card
func BuildCommittedCard(txn *models.QueryTransaction) *chat.CardWithId {
	return &chat.CardWithId{
		CardId: "transaction_committed_" + txn.ID.String(),
		Card: &chat.GoogleAppsCardV1Card{
			Header: &chat.GoogleAppsCardV1CardHeader{
				Title:    "✅ Transaction Committed",
				Subtitle: "QueryBase",
			},
			Sections: []*chat.GoogleAppsCardV1Section{
				{
					Widgets: []*chat.GoogleAppsCardV1Widget{
						decoratedText("Status", "Committed"),
						decoratedText("Affected Rows", fmt.Sprintf("%d", txn.AffectedRows)),
						decoratedText("Completed At", formatTimestamp(time.Now())),
					},
				},
			},
		},
	}
}

// BuildRolledBackCard creates the "Transaction Rolled Back" card
func BuildRolledBackCard(txn *models.QueryTransaction) *chat.CardWithId {
	return &chat.CardWithId{
		CardId: "transaction_rolledback_" + txn.ID.String(),
		Card: &chat.GoogleAppsCardV1Card{
			Header: &chat.GoogleAppsCardV1CardHeader{
				Title:    "🔄 Transaction Rolled Back",
				Subtitle: "QueryBase",
			},
			Sections: []*chat.GoogleAppsCardV1Section{
				{
					Widgets: []*chat.GoogleAppsCardV1Widget{
						decoratedText("Status", "Rolled Back"),
						decoratedText("Completed At", formatTimestamp(time.Now())),
					},
				},
			},
		},
	}
}

// BuildErrorCard creates an error notification card
func BuildErrorCard(title, message string) *chat.CardWithId {
	return &chat.CardWithId{
		CardId: "error",
		Card: &chat.GoogleAppsCardV1Card{
			Header: &chat.GoogleAppsCardV1CardHeader{
				Title:    "⚠️ " + title,
				Subtitle: "QueryBase",
			},
			Sections: []*chat.GoogleAppsCardV1Section{
				{
					Widgets: []*chat.GoogleAppsCardV1Widget{
						{
							TextParagraph: &chat.GoogleAppsCardV1TextParagraph{
								Text: message,
							},
						},
					},
				},
			},
		},
	}
}

// --- Helper functions ---

// decoratedText creates a key-value decorated text widget (GPOS DBHUB style)
func decoratedText(label, content string) *chat.GoogleAppsCardV1Widget {
	return &chat.GoogleAppsCardV1Widget{
		DecoratedText: &chat.GoogleAppsCardV1DecoratedText{
			TopLabel: label,
			Text:     content,
		},
	}
}

// actionButton creates an interactive button with an action callback
func actionButton(text, actionName, paramValue, color string) *chat.GoogleAppsCardV1Button {
	return &chat.GoogleAppsCardV1Button{
		Text: text,
		OnClick: &chat.GoogleAppsCardV1OnClick{
			Action: &chat.GoogleAppsCardV1Action{
				Function: actionName,
				Parameters: []*chat.GoogleAppsCardV1ActionParameter{
					{
						Key:   "id",
						Value: paramValue,
					},
				},
			},
		},
		Color: parseColor(color),
	}
}

// linkButton creates a button that opens a URL
func linkButton(text, url string) *chat.GoogleAppsCardV1Button {
	return &chat.GoogleAppsCardV1Button{
		Text: text,
		OnClick: &chat.GoogleAppsCardV1OnClick{
			OpenLink: &chat.GoogleAppsCardV1OpenLink{
				Url: url,
			},
		},
	}
}

// parseColor converts hex color to Google Chat Color struct
func parseColor(hex string) *chat.Color {
	if len(hex) != 7 || hex[0] != '#' {
		return nil
	}
	r := hexToFloat(hex[1:3])
	g := hexToFloat(hex[3:5])
	b := hexToFloat(hex[5:7])
	return &chat.Color{Red: r, Green: g, Blue: b}
}

func hexToFloat(h string) float64 {
	var val int
	for _, c := range h {
		val *= 16
		if c >= '0' && c <= '9' {
			val += int(c - '0')
		} else if c >= 'a' && c <= 'f' {
			val += int(c-'a') + 10
		} else if c >= 'A' && c <= 'F' {
			val += int(c-'A') + 10
		}
	}
	return float64(val) / 255.0
}

// operationIcon returns an icon for the operation type
func operationIcon(opType models.OperationType) string {
	switch opType {
	case "INSERT":
		return "➕"
	case "UPDATE":
		return "✏️"
	case "DELETE":
		return "🗑️"
	default:
		return "📝"
	}
}

// formatTimestamp formats a time for display
func formatTimestamp(t time.Time) string {
	return t.Format("1/2/2006, 3:04:05 PM")
}

// shortUUID returns first 8 chars of a UUID for display
func shortUUID(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}

// truncateSQL truncates a SQL query for card display
func truncateSQL(sql string, maxLen int) string {
	if len(sql) <= maxLen {
		return sql
	}
	return sql[:maxLen] + "..."
}

// buildPreviewDataSection creates a section showing preview data
func buildPreviewDataSection(txn *models.QueryTransaction) *chat.GoogleAppsCardV1Section {
	if txn.PreviewData == "" || txn.PreviewData == "null" {
		return &chat.GoogleAppsCardV1Section{
			Header: "Preview Data",
			Widgets: []*chat.GoogleAppsCardV1Widget{
				{
					TextParagraph: &chat.GoogleAppsCardV1TextParagraph{
						Text: "<i>No preview data available</i>",
					},
				},
			},
		}
	}

	return &chat.GoogleAppsCardV1Section{
		Header: "Preview Data",
		Widgets: []*chat.GoogleAppsCardV1Widget{
			{
				TextParagraph: &chat.GoogleAppsCardV1TextParagraph{
					Text: fmt.Sprintf("<code>%s</code>", truncateSQL(txn.PreviewData, 1000)),
				},
			},
		},
	}
}
