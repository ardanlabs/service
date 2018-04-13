package handlers

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
)

// Span represents the API to collect span data.
type Span struct {
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// Publish takes a batch and publishes that to a host system.
func (s *Span) Publish(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	// var usr user.CreateUser
	// if err := web.Unmarshal(r.Body, &usr); err != nil {
	// 	return errors.Wrap(err, "")
	// }

	// nUsr, err := user.Create(ctx, dbConn, &usr)
	// if err = check(err); err != nil {
	// 	return errors.Wrapf(err, "User: %+v", &usr)
	// }

	// web.Respond(ctx, w, nUsr, http.StatusCreated)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("*********>", err)
		return err
	}
	log.Println("*********>", string(data))

	return nil
}
