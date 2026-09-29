package report

import "strings"

type qixinRelationState uint8

const (
	qixinRelationInsufficient qixinRelationState = iota
	qixinRelationNotFound
	qixinRelationShareholder
	qixinRelationBranch
	qixinRelationCounterpartyShareholder
	qixinRelationCounterpartyBranch
)

func (state qixinRelationState) confirmed() bool {
	return state != qixinRelationInsufficient && state != qixinRelationNotFound
}

func (state qixinRelationState) summary() string {
	switch state {
	case qixinRelationShareholder:
		return "企业直接股东"
	case qixinRelationBranch:
		return "企业分支机构"
	case qixinRelationCounterpartyShareholder:
		return "企业直接持股的交易对手方"
	case qixinRelationCounterpartyBranch:
		return "企业为交易对手方分支机构"
	case qixinRelationNotFound:
		return "已核验直接股东及分支机构范围内未发现关系"
	default:
		return "关系核验覆盖不足"
	}
}

type qixinRelationEvidence struct {
	state      qixinRelationState
	share      float64
	shareKnown bool
	changes    []string
	refs       []string
}

func (relation qixinRelationEvidence) summary() string {
	text := relation.state.summary()
	if relation.state == qixinRelationCounterpartyShareholder && relation.shareKnown {
		switch {
		case relation.share == 100:
			text = "企业全资子公司（企业持股100.00%）"
		case relation.share > 50:
			text = "企业控股子公司（企业持股" + qixinOwnershipPercentText(relation.share, true) + "）"
		default:
			text += "（企业持股" + qixinOwnershipPercentText(relation.share, true) + "）"
		}
	}
	if len(relation.changes) > 0 {
		text += "；工商股东变更记录：" + strings.Join(relation.changes, "；")
	}
	return text
}

// Identifiers take precedence. A full-name fallback is allowed only when one
// side lacks an identifier; an explicit identifier conflict cannot match.
func qixinRelationIdentity(name, id, expectedName, expectedID string) bool {
	id, expectedID = qixinIdentityText(id), qixinIdentityText(expectedID)
	if id != "" && id != "-" && expectedID != "" && expectedID != "-" {
		return id == expectedID
	}
	return qixinIdentityText(name) != "" && qixinIdentityText(name) == qixinIdentityText(expectedName)
}

func qixinRelationCollection(detail any, collection string) []map[string]any {
	value, _, _ := findFirst(detail, collection)
	return findAllMaps(value, func(row map[string]any) bool { return mapText(row, "Name") != "" })
}

// Read only Partners/Branches, never names embedded in employees or historical
// changes. Both sides must come from the task's verified subject collection.
func qixinInspectRelation(ctx *reportContext, item counterpartyItem) qixinRelationEvidence {
	own := qixinOwnSubject(ctx)
	r := qixinRelationEvidence{state: qixinRelationInsufficient, refs: []string{"qixin_external/subjects[role=企业自身]/detail"}}
	if own == nil || qixinIdentityText(item.Name) == "" && qixinIdentityText(item.TaxID) == "" {
		return r
	}
	detail := qixinSubjectData(own, qixinDetailField)
	for _, partner := range qixinRelationCollection(detail, "Partners") {
		if qixinRelationIdentity(mapText(partner, "Name"), mapText(partner, "IdentifyNo", "CreditNo"), item.Name, item.TaxID) {
			r.state = qixinRelationShareholder
			return r
		}
	}
	for _, branch := range qixinRelationCollection(detail, "Branches") {
		if qixinRelationIdentity(mapText(branch, "Name"), mapText(branch, "CreditNo"), item.Name, item.TaxID) {
			r.state = qixinRelationBranch
			return r
		}
	}
	for _, role := range []string{qixinImportantCustomerRole, qixinImportantSupplierRole} {
		subject := qixinCounterpartySubject(ctx, item, role)
		if subject == nil {
			continue
		}
		otherDetail := qixinSubjectData(subject, qixinDetailField)
		ref := "qixin_external/subjects[taxId=" + mapText(subject, "taxId") + ";name=" + mapText(subject, "name") + "]/detail"
		for _, partner := range qixinRelationCollection(otherDetail, "Partners") {
			if !qixinRelationIdentity(mapText(partner, "Name"), mapText(partner, "IdentifyNo", "CreditNo"), mapText(own, "name"), mapText(own, "taxId")) {
				continue
			}
			r.share, r.shareKnown = qixinOwnershipPercent(mapText(partner, "StockPercent"))
			if r.shareKnown && (r.share <= 0 || r.share > 100) {
				continue
			}
			r.state = qixinRelationCounterpartyShareholder
			r.refs = append(r.refs, ref+"/Partners")
			// Preserve registry change dates as evidence, not as an inferred period
			// of control. Current ownership alone cannot describe the whole year.
			changes, _, _ := findFirst(otherDetail, "Changerecords")
			for _, change := range findAllMaps(changes, func(row map[string]any) bool { return containsAny(mapText(row, "ChangeItem"), "股东", "投资人") }) {
				r.changes = append(r.changes, mapText(change, "ChangeDate")+"："+mapText(change, "BeforeContent")+" → "+mapText(change, "AfterContent"))
			}
			if len(r.changes) > 0 {
				r.refs = append(r.refs, ref+"/Changerecords")
			}
			return r
		}
		for _, branch := range qixinRelationCollection(otherDetail, "Branches") {
			if qixinRelationIdentity(mapText(branch, "Name"), mapText(branch, "CreditNo"), mapText(own, "name"), mapText(own, "taxId")) {
				r.state = qixinRelationCounterpartyBranch
				r.refs = append(r.refs, ref+"/Branches")
				return r
			}
		}
	}
	_, _, partnersPresent := findFirst(detail, "Partners")
	_, _, branchesPresent := findFirst(detail, "Branches")
	if partnersPresent && branchesPresent {
		r.state = qixinRelationNotFound
	}
	return r
}

func qixinDirectRelation(ctx *reportContext, item counterpartyItem) qixinRelationState {
	return qixinInspectRelation(ctx, item).state
}

func qixinVerifiedRelationSummary(ctx *reportContext, item counterpartyItem) string {
	return qixinInspectRelation(ctx, item).summary()
}

func addCounterpartyRelationEvidence(d *SectionAnalysisData, ctx *reportContext, item counterpartyItem, prefix string) string {
	r := qixinInspectRelation(ctx, item)
	id := d.text(prefix+".relation", "已核验直接关系范围", "relationship", r.summary(), r.refs...)
	if len(r.changes) > 0 {
		d.text(prefix+".relation-changes", "交易对手方股东变更记录", "relationship", strings.Join(r.changes, "；"), r.refs...)
	}
	return id
}
