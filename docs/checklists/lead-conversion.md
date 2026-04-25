### Lead Conversion

**Go Backend**
- [x] POST /leads/{id}/convert endpoint
- [x] Find or create account (by ID or by name)
- [x] Create opportunity linked to account (account_opportunities join)
- [x] Create contact from lead data, linked to account + opportunity
- [x] Set lead status to "converted"
- [x] Update counter caches (account contacts_count, opportunities_count, campaign opportunities_count)
- [x] Audit trail (version records for all created entities)
- [x] All in a single DB transaction
- [x] Route registration
- [x] Tests: 11 tests covering success (new/existing/by-name account), counter caches, campaign inheritance, already converted, not found, forbidden, validation errors, nonexistent account, no auth

**React Frontend**
- [x] Conversion form on lead detail page (ConvertLeadForm component)
- [x] Account selector (existing dropdown or new name input)
- [x] Opportunity fields (name, stage, amount, probability, close date)
- [x] Access level selector
- [x] Success redirect to contact detail page
- [x] "Converted" badge shown instead of Convert button for converted leads

**Verified**
- [x] Counter caches correct after conversion
- [x] Audit trail records created
- [x] Lead data copied to contact (name, title, email, phone, mobile, background_info, do_not_call)
- [x] Campaign ID inherited by opportunity
