package crystal

import "prismarine/shard/remote"

var _ remote.Client = (*client)(nil)

type client struct{}
