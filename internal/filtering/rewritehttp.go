package filtering

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/golibs/log"
)

// TODO(d.kolyshev): Use [rewrite.Item] instead.
type rewriteEntryJSON struct {
	Domain string `json:"domain"`
	Answer string `json:"answer"`
}

// handleRewriteList is the handler for the GET /control/rewrite/list HTTP API.
func (d *DNSFilter) handleRewriteList(w http.ResponseWriter, r *http.Request) {
	arr := []*rewriteEntryJSON{}

	// 定制
	p := r.URL.Query().Get("param")
	// 定制

	func() {
		d.confMu.RLock()
		defer d.confMu.RUnlock()

		for _, ent := range d.conf.Rewrites {
			// 定制
			// jsonEnt := rewriteEntryJSON{
			// 	Domain: ent.Domain,
			// 	Answer: ent.Answer,
			// }
			// arr = append(arr, &jsonEnt)
			if strings.Contains(ent.Domain, p) || strings.Contains(ent.Answer, p) {
				jsonEnt := rewriteEntryJSON{
					Domain: ent.Domain,
					Answer: ent.Answer,
				}
				arr = append(arr, &jsonEnt)
			}
			// 定制
		}
	}()

	aghhttp.WriteJSONResponseOK(w, r, arr)
}

// handleRewriteAdd is the handler for the POST /control/rewrite/add HTTP API.
func (d *DNSFilter) handleRewriteAdd(w http.ResponseWriter, r *http.Request) {
	rwJSON := rewriteEntryJSON{}
	err := json.NewDecoder(r.Body).Decode(&rwJSON)
	if err != nil {
		aghhttp.Error(r, w, http.StatusBadRequest, "json.Decode: %s", err)

		return
	}

	/* 定制 */
	// 校验主机记录是否重复
	err = checkForDuplicateDomains(d, rwJSON.Domain, rwJSON.Answer, err)
	if err != nil {
		aghhttp.Error(r, w, http.StatusBadRequest, "DNS 解析 | 您当前 添加 的主机记录 %s", err)
		return
	}
	/* 定制 */

	rw := &LegacyRewrite{
		Domain: rwJSON.Domain,
		Answer: rwJSON.Answer,
	}

	err = rw.normalize()
	if err != nil {
		// Shouldn't happen currently, since normalize only returns a non-nil
		// error when a rewrite is nil, but be change-proof.
		aghhttp.Error(r, w, http.StatusBadRequest, "normalizing: %s", err)

		return
	}

	func() {
		d.confMu.Lock()
		defer d.confMu.Unlock()

		d.conf.Rewrites = append(d.conf.Rewrites, rw)
		log.Debug(
			"rewrite: added element: %s -> %s [%d]",
			rw.Domain,
			rw.Answer,
			len(d.conf.Rewrites),
		)
	}()

	d.conf.ConfigModified()
}

// handleRewriteDelete is the handler for the POST /control/rewrite/delete HTTP
// API.
func (d *DNSFilter) handleRewriteDelete(w http.ResponseWriter, r *http.Request) {
	jsent := rewriteEntryJSON{}
	err := json.NewDecoder(r.Body).Decode(&jsent)
	if err != nil {
		aghhttp.Error(r, w, http.StatusBadRequest, "json.Decode: %s", err)

		return
	}

	entDel := &LegacyRewrite{
		Domain: jsent.Domain,
		Answer: jsent.Answer,
	}
	arr := []*LegacyRewrite{}

	func() {
		d.confMu.Lock()
		defer d.confMu.Unlock()

		for _, ent := range d.conf.Rewrites {
			if ent.equal(entDel) {
				log.Debug("rewrite: removed element: %s -> %s", ent.Domain, ent.Answer)

				continue
			}

			arr = append(arr, ent)
		}
		d.conf.Rewrites = arr
	}()

	d.conf.ConfigModified()
}

// rewriteUpdateJSON is a struct for JSON object with rewrite rule update info.
type rewriteUpdateJSON struct {
	Target rewriteEntryJSON `json:"target"`
	Update rewriteEntryJSON `json:"update"`
}

// handleRewriteUpdate is the handler for the PUT /control/rewrite/update HTTP
// API.
func (d *DNSFilter) handleRewriteUpdate(w http.ResponseWriter, r *http.Request) {
	updateJSON := rewriteUpdateJSON{}
	err := json.NewDecoder(r.Body).Decode(&updateJSON)
	if err != nil {
		aghhttp.Error(r, w, http.StatusBadRequest, "json.Decode: %s", err)

		return
	}

	/* 定制 */
	// 校验主机记录是否重复
	/*if updateJSON.Update.Domain != updateJSON.Target.Domain {

	  }*/
	err = checkForDuplicateDomains(d, updateJSON.Update.Domain, updateJSON.Update.Answer, err)
	if err != nil {
		aghhttp.Error(r, w, http.StatusBadRequest, "DNS 解析    您当前 修改 的主机记录 %s", err)
		return
	}
	/* 定制 */

	rwDel := &LegacyRewrite{
		Domain: updateJSON.Target.Domain,
		Answer: updateJSON.Target.Answer,
	}

	rwAdd := &LegacyRewrite{
		Domain: updateJSON.Update.Domain,
		Answer: updateJSON.Update.Answer,
	}

	err = rwAdd.normalize()
	if err != nil {
		// Shouldn't happen currently, since normalize only returns a non-nil
		// error when a rewrite is nil, but be change-proof.
		aghhttp.Error(r, w, http.StatusBadRequest, "normalizing: %s", err)

		return
	}

	index := -1
	defer func() {
		if index >= 0 {
			d.conf.ConfigModified()
		}
	}()

	d.confMu.Lock()
	defer d.confMu.Unlock()

	index = slices.IndexFunc(d.conf.Rewrites, rwDel.equal)
	if index == -1 {
		aghhttp.Error(r, w, http.StatusBadRequest, "target rule not found")

		return
	}

	d.conf.Rewrites = slices.Replace(d.conf.Rewrites, index, index+1, rwAdd)

	log.Debug("rewrite: removed element: %s -> %s", rwDel.Domain, rwDel.Answer)
	log.Debug("rewrite: added element: %s -> %s", rwAdd.Domain, rwAdd.Answer)
}

/* 定制 */
// 校验主机记录是否重复
func checkForDuplicateDomains(d *DNSFilter, domain string, answer string, err error) error {
	num := 0
	d.confMu.RLock()
	defer d.confMu.RUnlock()
	for _, ent := range d.conf.Rewrites {
		if ent.Domain == domain && ent.Answer == answer {
			num += 1
		}
	}

	if num == 0 {
		err = nil
	} else {
		err = errors.New("已经存在")
	}
	return err
}

/* 定制 */
