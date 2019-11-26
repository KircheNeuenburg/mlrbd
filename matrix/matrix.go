package matrix

import (
	"log"
	"maunium.net/go/mautrix"
	"mlrbd/config"
)

var (
	m *mautrix.Client
	c *config.Config
)

func StartMatrix(conf *config.Config) {
	if conf != nil {
		c = conf
	}
	var err error
	m, err = mautrix.NewClient("https://"+c.Matrix.Homeserver, c.Matrix.Mxid, c.Matrix.AccessToken)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Matrix started")
}

func CreateMatrixRoom(n string) string {
	req := mautrix.ReqCreateRoom{
		Visibility: "private",
		Name:       n + c.Matrix.RoomSuffix,
		Preset:     "private_chat",
	}
	resp, err := m.CreateRoom(&req)
	if err != nil {
		log.Fatal(err)
	}
	return resp.RoomID
}

func DeleteMatrixRoom(rid string) {
	u := m.BuildURL("rooms", rid, "members")
	resp := struct {
		Chunk []struct {
			Sender  string `json:"sender"`
			UserId  string `json:"state_key"`
			Content struct {
				Membership string `json:"membership"`
			} `json:"content"`
		}
	}{}
	_, err := m.MakeRequest("GET", u, nil, &resp)
	if err != nil {
		log.Fatal(err)
	}
	for _, u := range resp.Chunk {
		if u.UserId != c.Matrix.Mxid && (u.Content.Membership == "join" || u.Content.Membership == "invite") {
			req := mautrix.ReqKickUser{UserID: u.UserId, Reason: c.Matrix.KickMessage}
			_, err := m.KickUser(rid, &req)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	m.LeaveRoom(rid)
}

func MatrixUsers(rid string) (mu []string) {
	u := m.BuildURL("rooms", rid, "members")
	resp := struct {
		Chunk []struct {
			Sender  string `json:"sender"`
			UserId  string `json:"state_key"`
			Content struct {
				Membership string `json:"membership"`
			} `json:"content"`
		}
	}{}
	_, err := m.MakeRequest("GET", u, nil, &resp)
	if err != nil {
		log.Fatal(err)
	}
	for _, u := range resp.Chunk {
		if u.UserId != c.Matrix.Mxid && (u.Content.Membership == "join" || u.Content.Membership == "invite") {
			mu = append(mu, u.UserId)
		}
	}
	return
}

func HandleCreatedUsers(rid string, mu []string) {
	for _, u := range mu {
		req := mautrix.ReqInviteUser{UserID: u}
		_, err := m.InviteUser(rid, &req)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Invited user ", u, " to room ", rid)
	}

	return
}
func HandleRemovedUsers(rid string, mu []string) {
	for _, u := range mu {
		req := mautrix.ReqKickUser{UserID: u, Reason: c.Matrix.KickMessage}
		_, err := m.KickUser(rid, &req)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Removed user ", u, " from room ", rid)
	}

	return
}

func EnableEncryption(rid string) {
	u := m.BuildURL("rooms", rid, "state", "m.room.encryption")
	req := struct {
		Algorithm string `json:"algorithm"`
	}{}
	req.Algorithm = "m.megolm.v1.aes-sha2"
	_, err := m.MakeRequest("PUT", u, &req, nil)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func SetRoomName(rid string, n string) {
	u := m.BuildURL("rooms", rid, "state", "m.room.name")
	req := struct {
		Name string `json:"name"`
	}{}
	req.Name = n + c.Matrix.RoomSuffix
	_, err := m.MakeRequest("PUT", u, &req, nil)
	if err != nil {
		log.Fatal(err)
	}
	return
}
func MatrixRooms() (mr []string, err error) {
	resp, err := m.JoinedRooms()
	if err != nil {
		log.Fatal(err)
	}
	mr = resp.JoinedRooms
	return
}
