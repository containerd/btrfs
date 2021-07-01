/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package btrfs

import (
	"encoding/binary"
	"fmt"
	"os"
)

func lookupUUIDSubvolItem(f *os.File, uuid UUID) (objectID, error) {
	return uuidTreeLookupAny(f, uuid, uuidKeySubvol)
}

func lookupUUIDReceivedSubvolItem(f *os.File, uuid UUID) (objectID, error) {
	return uuidTreeLookupAny(f, uuid, uuidKeyReceivedSubvol)
}

func (id UUID) toKey() (objID objectID, off uint64) {
	objID = objectID(binary.LittleEndian.Uint64(id[:8]))
	off = binary.LittleEndian.Uint64(id[8:16])
	return
}

// uuidTreeLookupAny searches uuid tree for a given uuid in specified field.
// It returns ErrNotFound if object was not found.
func uuidTreeLookupAny(f *os.File, uuid UUID, typ treeKeyType) (objectID, error) {
	objId, off := uuid.toKey()
	args := btrfs_ioctl_search_key{
		tree_id:      uuidTreeObjectid,
		min_objectid: objId,
		max_objectid: objId,
		min_type:     typ,
		max_type:     typ,
		min_offset:   off,
		max_offset:   off,
		max_transid:  maxUint64,
		nr_items:     1,
	}
	res, err := treeSearchRaw(f, args)
	if err != nil {
		return 0, err
	} else if len(res) < 1 {
		return 0, ErrNotFound
	}
	out := res[0]
	if len(out.Data) != 8 {
		return 0, fmt.Errorf("btrfs: uuid item with illegal size %d", len(out.Data))
	}
	return objectID(binary.LittleEndian.Uint64(out.Data)), nil
}
