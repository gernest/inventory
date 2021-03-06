package inventory

import (
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/fritz0705/inventory/reichelt"
)

var APIs = map[string]func(string)string{
	"reichelt": reichelt.ResolveID,
}

func (app *Application) DistributorPartRedirect(w http.ResponseWriter, r *http.Request) {
	tx := app.DB.MustBegin()
	defer tx.Rollback()

	id := path.Base(r.URL.Path)

	distributorPartView := new(DistributorPartView)
	err := tx.Get(distributorPartView, `SELECT * FROM 'distributor_part_view'
	WHERE "id" = ?`, id)
	if app.SQLError(w, r, err) {
		return
	}

	tx.Commit()

	http.Redirect(w, r, distributorPartView.PartURL(), http.StatusMovedPermanently)
}

func (app *Application) DeleteDistributorPart(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)

	tx := app.DB.MustBegin()
	defer tx.Rollback()

	distributorPart := new(DistributorPart)
	err := tx.Get(distributorPart, `SELECT * FROM 'distributor_part' WHERE "id" = ?`, id)
	if app.SQLError(w, r, err) {
		return
	}

	_, err = tx.Exec(`DELETE FROM 'distributor_part' WHERE "id" = ?`, id)
	app.SQLError(w, r, err)

	tx.Commit()

	http.Redirect(w, r, fmt.Sprintf("/parts/edit/%d", distributorPart.PartId), http.StatusSeeOther)
}

func (app *Application) CreateDistributorPart(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.Error(w, err)
		return
	}

	tx := app.DB.MustBegin()
	defer tx.Rollback()

	id, _ := strconv.Atoi(path.Base(r.URL.Path))
	if id == 0 {
		app.NotFoundHandler(w, r)
		return
	}

	distributorId, _ := strconv.Atoi(r.PostForm.Get("distributor"))

	distributor := new(Distributor)
	err = tx.Get(distributor, `SELECT * FROM 'distributor' WHERE "id" = ?`, distributorId)
	if app.SQLError(w, r, err) {
		return
	}

	part := new(Part)
	err = tx.Get(part, `SELECT * FROM 'part' WHERE "id" = ?`, id)
	if app.SQLError(w, r, err) {
		return
	}

	distributorPart := &DistributorPart{
		PartId: int64(id),
	}

	err = distributorPart.LoadForm(r.PostForm)
	if err != nil {
		app.Error(w, err)
		return
	}

	if distributor.Api.Valid {
		if api := APIs[distributor.Api.String]; api != nil {
			distributorPart.Key = api(distributorPart.Key)
		}
	}

	err = distributorPart.Save(tx)
	if err != nil {
		app.Error(w, err)
		return
	}

	tx.Commit()

	http.Redirect(w, r, fmt.Sprintf("/parts/edit/%d", id), http.StatusSeeOther)
}
