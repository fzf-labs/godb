import (
    "context"
	"errors"
	"github.com/fzf-labs/godb/orm/condition"
	"github.com/fzf-labs/godb/orm/dbcache"
	"github.com/fzf-labs/godb/orm/encoding"
	"github.com/fzf-labs/godb/orm/gen/config"
    "{{.daoPkgPath}}"
    "{{.modelPkgPath}}"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)
