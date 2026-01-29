# Approval Workflow Guide

Learn how to use QueryBase's approval workflow for write operations.

## Overview

QueryBase requires approval for all write operations (INSERT, UPDATE, DELETE, DDL) to ensure data safety and provide audit trails.

## How It Works

### For Users Submitting Queries

1. **Write Your Query**
   - Open the SQL Editor
   - Write your INSERT, UPDATE, DELETE, or DDL query
   - Click "Execute"

2. **Approval Request Created**
   - System detects write operation
   - Creates approval request automatically
   - Notifies eligible approvers via Google Chat
   - Query status: "Pending Approval"

3. **Wait for Approval**
   - Monitor approval status in Approvals dashboard
   - You'll receive notification when decision is made
   - Can add comments to provide context

4. **Execution**
   - If approved: Background worker executes query
   - Results appear in Query History
   - You receive completion notification
   - If rejected: You receive notification with reason

### For Approvers

1. **Receive Notification**
   - Google Chat webhook notification
   - Contains query text, operation type, requester
   - Link to approval dashboard

2. **Review Query**
   - Open Approvals dashboard
   - Review query text and execution plan
   - Check requester's permissions
   - Read any comments from requester

3. **Make Decision**
   - **Approve**: Query executes immediately
   - **Reject**: Query is cancelled
   - Add comments explaining decision

4. **Transaction Management (Optional)**
   - For approved queries, preview results first
   - Commit to apply changes
   - Rollback to undo changes

## Transaction Preview

For certain write operations, you can preview results before committing:

1. Approve the query
2. Background worker starts transaction
3. Preview affected rows
4. Choose:
   - **Commit**: Apply changes permanently
   - **Rollback**: Undo all changes

## Permission Levels

### `can_read`
- Execute SELECT queries only
- Cannot submit write operations

### `can_write`
- Submit write operation requests
- Cannot approve queries

### `can_approve`
- Review and approve/reject write operations
- Execute SELECT queries
- Submit write operation requests

## Approval Dashboard

### Features

- **Pending Approvals**: List of queries awaiting review
- **Recent Decisions**: Your approval history
- **Query Details**: Full query text, requester, timestamp
- **Comments**: Add context or feedback
- **Transaction Controls**: Commit/rollback for transaction previews

### Filtering

- Filter by status (Pending, Approved, Rejected)
- Filter by data source
- Filter by requester
- Sort by date

## Best Practices

### For Requesters

1. **Add Context**
   - Use comments to explain WHY you need to make changes
   - Include business justification
   - Mention any risks

2. **Test First**
   - Use SELECT to verify affected rows
   - Use EXPLAIN to check performance
   - Use Dry Run for DELETE operations

3. **Start Small**
   - Break large operations into smaller batches
   - Easier to review and approve
   - Safer to execute

### For Approvers

1. **Review Carefully**
   - Check query syntax and logic
   - Verify WHERE clauses
   - Look for missing conditions

2. **Ask Questions**
   - If something seems unclear, ask for clarification
   - Use comments to communicate
   - Request changes before approving

3. **Document Decisions**
   - Always add comments for approvals/rejections
   - Explain reasoning
   - Creates audit trail

## Notification Timing

- **Immediate**: Google Chat webhook when request created
- **Immediate**: Notification when decision made
- **Immediate**: Notification when execution completes
- **Background**: Email summary (optional, to be implemented)

## Troubleshooting

### Query Stuck in "Pending"

**Possible causes:**
- No eligible approvers for data source
- Approvers are unavailable
- Notification webhook failed

**Solutions:**
- Check data source permissions
- Contact admin to assign approvers
- Check notification configuration

### Approval Failed

**Possible causes:**
- Query syntax error
- Permission denied on data source
- Constraint violation
- Transaction timeout

**Solutions:**
- Check error message
- Verify query syntax
- Check data source permissions
- Review constraints

### Transaction Rollback Failed

**Possible causes:**
- Transaction already committed
- Connection lost to data source
- Timeout exceeded

**Solutions:**
- Check connection status
- Verify transaction state
- Contact administrator

## Related Features

- **[Query Features](query-features.md)** - EXPLAIN and Dry Run
- **[Schema Browser](schema-browser.md)** - Explore before modifying
- **[Query History](query-features.md)** - Track all executions

---

**Need help?** Check the [main documentation](../README.md) or [contact support](../README.md).
