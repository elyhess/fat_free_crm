#!/usr/bin/env bash
#
# Smoke test: verify every major Go API endpoint against a live database.
# Requires: server running on localhost:8080, demo data loaded.
#
# Usage: ./scripts/smoke-test.sh
#
set -euo pipefail

BASE="http://localhost:8080/api/v1"
PASS=0
FAIL=0
ERRORS=""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m'

check() {
  local desc="$1"
  local method="$2"
  local url="$3"
  local expected_status="${4:-200}"
  local body="${5:-}"

  local args=(-s -o /tmp/smoke-response.json -w "%{http_code}" -X "$method")
  if [ -n "$TOKEN" ]; then
    args+=(-H "Authorization: Bearer $TOKEN")
  fi
  args+=(-H "Content-Type: application/json")
  if [ -n "$body" ]; then
    args+=(-d "$body")
  fi
  args+=("${BASE}${url}")

  local status
  status=$(curl "${args[@]}" 2>/dev/null || echo "000")

  if [ "$status" = "$expected_status" ]; then
    echo -e "  ${GREEN}PASS${NC} [$status] $desc"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}FAIL${NC} [$status] $desc (expected $expected_status)"
    local resp
    resp=$(cat /tmp/smoke-response.json 2>/dev/null | head -c 200)
    echo -e "       Response: $resp"
    FAIL=$((FAIL + 1))
    ERRORS="${ERRORS}\n  - $desc: got $status, expected $expected_status"
  fi
}

# Helper: extract JSON field
json_field() {
  python3 -c "import json,sys; print(json.load(sys.stdin)$1)" < /tmp/smoke-response.json 2>/dev/null
}

echo ""
echo "=================================="
echo " Fat Free CRM — API Smoke Tests"
echo "=================================="
echo ""

# ─── Authentication ─────────────────────────────────────────────
echo -e "${YELLOW}Authentication${NC}"
TOKEN=""

# Login with admin user
check "Login (valid credentials)" POST "/auth/login" 200 '{"login":"admin","password":"Dem0P@ssword!!"}'
TOKEN=$(json_field "['token']" 2>/dev/null || echo "")
if [ -z "$TOKEN" ]; then
  echo -e "  ${RED}FATAL: Could not obtain auth token. Aborting.${NC}"
  exit 1
fi
echo -e "  Token obtained: ${TOKEN:0:20}..."

# Login with bad password
check "Login (invalid password)" POST "/auth/login" 401 '{"login":"admin","password":"wrong"}'

# ─── Profile ────────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Profile${NC}"
check "Get profile" GET "/profile" 200
check "List users" GET "/users" 200

# ─── Accounts ───────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Accounts${NC}"
check "List accounts" GET "/accounts?page=1&per_page=5" 200
ACCOUNT_COUNT=$(json_field "['total']" 2>/dev/null || echo "0")
echo -e "  Total accounts: $ACCOUNT_COUNT"

check "Get account #1" GET "/accounts/1" 200
check "Account contacts" GET "/accounts/1/contacts" 200
check "Account opportunities" GET "/accounts/1/opportunities" 200
check "Account comments" GET "/accounts/1/comments" 200
check "Account tags" GET "/accounts/1/tags" 200
check "Account addresses" GET "/accounts/1/addresses" 200
check "Account versions" GET "/accounts/1/versions" 200
check "Account emails" GET "/accounts/1/emails" 200
check "Account custom fields" GET "/accounts/1/custom_fields" 200
check "Account subscription" GET "/accounts/1/subscription" 200
check "Accounts autocomplete" GET "/accounts/autocomplete?q=a" 200
check "Accounts export" GET "/accounts/export" 200

# ─── Contacts ───────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Contacts${NC}"
check "List contacts" GET "/contacts?page=1&per_page=5" 200
check "Get contact #1" GET "/contacts/1" 200
check "Contact opportunities" GET "/contacts/1/opportunities" 200
check "Contact comments" GET "/contacts/1/comments" 200
check "Contact versions" GET "/contacts/1/versions" 200
check "Contacts autocomplete" GET "/contacts/autocomplete?q=a" 200
check "Contacts export" GET "/contacts/export" 200
check "Contacts vCard export" GET "/contacts/export/vcard" 200

# ─── Leads ──────────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Leads${NC}"
check "List leads" GET "/leads?page=1&per_page=5" 200
check "Get lead #1" GET "/leads/1" 200
check "Lead comments" GET "/leads/1/comments" 200
check "Lead versions" GET "/leads/1/versions" 200
check "Leads autocomplete" GET "/leads/autocomplete?q=a" 200
check "Leads export" GET "/leads/export" 200

# ─── Opportunities ──────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Opportunities${NC}"
check "List opportunities" GET "/opportunities?page=1&per_page=5" 200
check "Get opportunity #1" GET "/opportunities/1" 200
check "Opportunity comments" GET "/opportunities/1/comments" 200
check "Opportunity versions" GET "/opportunities/1/versions" 200
check "Opportunities autocomplete" GET "/opportunities/autocomplete?q=a" 200
check "Opportunities export" GET "/opportunities/export" 200

# ─── Campaigns ──────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Campaigns${NC}"
check "List campaigns" GET "/campaigns?page=1&per_page=5" 200
check "Get campaign #1" GET "/campaigns/1" 200
check "Campaign leads" GET "/campaigns/1/leads" 200
check "Campaign opportunities" GET "/campaigns/1/opportunities" 200
check "Campaign comments" GET "/campaigns/1/comments" 200
check "Campaign versions" GET "/campaigns/1/versions" 200
check "Campaigns autocomplete" GET "/campaigns/autocomplete?q=a" 200
check "Campaigns export" GET "/campaigns/export" 200

# ─── Tasks ──────────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Tasks${NC}"
check "List tasks" GET "/tasks?page=1&per_page=5" 200
check "Get task #1" GET "/tasks/1" 200
check "Tasks export" GET "/tasks/export" 200

# ─── Dashboard ──────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Dashboard${NC}"
check "Task summary" GET "/dashboard/tasks" 200
check "Pipeline summary" GET "/dashboard/pipeline" 200
check "Activity feed" GET "/activity?limit=10" 200

# ─── Search ─────────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Search${NC}"
check "Search accounts" GET "/search?q=a&entity=accounts" 200
check "Search contacts" GET "/search?q=john&entity=contacts" 200
check "Search all" GET "/search?q=test" 200

# ─── Filtering ──────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Filtering${NC}"
check "Filter accounts by name" GET "/accounts?filter%5Bname_cont%5D=corp" 200
check "Filter leads by status" GET "/leads?filter%5Bstatus_eq%5D=new" 200
check "Filter opportunities by stage" GET "/opportunities?filter%5Bstage_eq%5D=prospecting" 200

# ─── Tags ───────────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Tags${NC}"
check "List all tags" GET "/tags" 200

# ─── Field Groups ───────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Field Groups${NC}"
check "List field groups (Account)" GET "/field_groups?entity=Account" 200

# ─── Admin ──────────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Admin${NC}"
check "Admin settings" GET "/admin/settings" 200
check "Admin groups" GET "/admin/groups" 200
check "Admin plugins" GET "/admin/plugins" 200
check "Admin research tools" GET "/admin/research_tools" 200

# ─── Saved Searches ─────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Saved Searches${NC}"
check "List saved searches" GET "/saved_searches" 200

# ─── Write Operations (Create → Update → Delete) ───────────────
echo ""
echo -e "${YELLOW}Write Operations${NC}"

# Create account
check "Create account" POST "/accounts" 201 '{"name":"Smoke Test Corp","access":"Public"}'
SMOKE_ACCOUNT_ID=$(json_field "['id']" 2>/dev/null || echo "0")
echo -e "  Created account #$SMOKE_ACCOUNT_ID"

# Update account
if [ "$SMOKE_ACCOUNT_ID" != "0" ]; then
  check "Update account" PUT "/accounts/$SMOKE_ACCOUNT_ID" 200 '{"name":"Smoke Test Corp Updated"}'
fi

# Create contact
check "Create contact" POST "/contacts" 201 '{"first_name":"Smoke","last_name":"Tester","access":"Public"}'
SMOKE_CONTACT_ID=$(json_field "['id']" 2>/dev/null || echo "0")
echo -e "  Created contact #$SMOKE_CONTACT_ID"

# Create lead
check "Create lead" POST "/leads" 201 '{"first_name":"Lead","last_name":"Smoker","status":"new","access":"Public"}'
SMOKE_LEAD_ID=$(json_field "['id']" 2>/dev/null || echo "0")
echo -e "  Created lead #$SMOKE_LEAD_ID"

# Reject lead
if [ "$SMOKE_LEAD_ID" != "0" ]; then
  check "Reject lead" PUT "/leads/$SMOKE_LEAD_ID/reject" 200
fi

# Create opportunity
check "Create opportunity" POST "/opportunities" 201 '{"name":"Smoke Deal","stage":"prospecting","amount":10000,"access":"Public"}'
SMOKE_OPP_ID=$(json_field "['id']" 2>/dev/null || echo "0")
echo -e "  Created opportunity #$SMOKE_OPP_ID"

# Create campaign
check "Create campaign" POST "/campaigns" 201 '{"name":"Smoke Campaign","status":"planned","access":"Public"}'
SMOKE_CAMPAIGN_ID=$(json_field "['id']" 2>/dev/null || echo "0")
echo -e "  Created campaign #$SMOKE_CAMPAIGN_ID"

# Create task
check "Create task" POST "/tasks" 201 '{"name":"Smoke Task","category":"call","bucket":"due_asap"}'
SMOKE_TASK_ID=$(json_field "['id']" 2>/dev/null || echo "0")
echo -e "  Created task #$SMOKE_TASK_ID"

# Complete / uncomplete task
if [ "$SMOKE_TASK_ID" != "0" ]; then
  check "Complete task" PUT "/tasks/$SMOKE_TASK_ID/complete" 200
  check "Uncomplete task" PUT "/tasks/$SMOKE_TASK_ID/uncomplete" 200
fi

# Add comment
if [ "$SMOKE_ACCOUNT_ID" != "0" ]; then
  check "Add comment" POST "/accounts/$SMOKE_ACCOUNT_ID/comments" 201 '{"comment":"Smoke test comment"}'
fi

# Add tag
if [ "$SMOKE_ACCOUNT_ID" != "0" ]; then
  check "Add tag" POST "/accounts/$SMOKE_ACCOUNT_ID/tags" 201 '{"name":"smoke-test"}'
fi

# Subscribe
if [ "$SMOKE_ACCOUNT_ID" != "0" ]; then
  check "Subscribe" POST "/accounts/$SMOKE_ACCOUNT_ID/subscribe" 200
  check "Get subscription" GET "/accounts/$SMOKE_ACCOUNT_ID/subscription" 200
  check "Unsubscribe" POST "/accounts/$SMOKE_ACCOUNT_ID/unsubscribe" 200
fi

# Create saved search
check "Create saved search" POST "/saved_searches" 201 '{"name":"Smoke Search","entity":"accounts","filters":{"name_cont":"smoke"}}'
SMOKE_SEARCH_ID=$(json_field "['id']" 2>/dev/null || echo "0")

# Clean up: delete created records
echo ""
echo -e "${YELLOW}Cleanup${NC}"
[ "$SMOKE_SEARCH_ID" != "0" ] && check "Delete saved search" DELETE "/saved_searches/$SMOKE_SEARCH_ID" 200
[ "$SMOKE_TASK_ID" != "0" ] && check "Delete task" DELETE "/tasks/$SMOKE_TASK_ID" 200
[ "$SMOKE_CAMPAIGN_ID" != "0" ] && check "Delete campaign" DELETE "/campaigns/$SMOKE_CAMPAIGN_ID" 200
[ "$SMOKE_OPP_ID" != "0" ] && check "Delete opportunity" DELETE "/opportunities/$SMOKE_OPP_ID" 200
[ "$SMOKE_LEAD_ID" != "0" ] && check "Delete lead" DELETE "/leads/$SMOKE_LEAD_ID" 200
[ "$SMOKE_CONTACT_ID" != "0" ] && check "Delete contact" DELETE "/contacts/$SMOKE_CONTACT_ID" 200
[ "$SMOKE_ACCOUNT_ID" != "0" ] && check "Delete account" DELETE "/accounts/$SMOKE_ACCOUNT_ID" 200

# ─── Import Template ────────────────────────────────────────────
echo ""
echo -e "${YELLOW}Import/Export${NC}"
check "Import template (accounts)" GET "/accounts/import/template" 200
check "Import template (contacts)" GET "/contacts/import/template" 200
check "Import template (leads)" GET "/leads/import/template" 200

# ─── Auth Flows (public) ───────────────────────────────────────
echo ""
echo -e "${YELLOW}Auth Flows (public, expect controlled failures)${NC}"
TOKEN_BAK="$TOKEN"
TOKEN=""
check "Forgot password (non-existent email — still 200)" POST "/auth/forgot-password" 200 '{"email":"nonexistent@example.com"}'
check "Reset password (bad token)" POST "/auth/reset-password" 422 '{"token":"badtoken","new_password":"Newpass123!!"}'
check "Confirm (bad token)" POST "/auth/confirm" 422 '{"token":"badtoken"}'
TOKEN="$TOKEN_BAK"

# ─── Summary ────────────────────────────────────────────────────
echo ""
echo "=================================="
echo -e " Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC}"
echo "=================================="

if [ $FAIL -gt 0 ]; then
  echo -e "\nFailed tests:${ERRORS}"
  exit 1
fi
