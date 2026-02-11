// Package presenters Utility methods to convert from models to protobuf messages
package presenters

import (
	"testbert/protobuf/collection"
	"testbert/server/model"
)

func Collection(in *model.Collection) *collection.Collection {
	return &collection.Collection{
		CollectionId:   in.ID.String(),
		CollectionData: in.Data,
		OrgView:        in.OrgView,
		OrgEdit:        in.OrgEdit,
		OrgShare:       in.OrgShare,
	}
}

func SharingToken(in *model.SharingToken) *collection.ShareToken {
	return &collection.ShareToken{
		CollectionId: in.CollectionID.String(),
		Token:        in.Token,
	}
}
