package infer_test

import "testing"

// A trait reading its own static properties through self:: or static::
// gets the types it declares for them.
func TestTraitStaticProperties(t *testing.T) {
	checkAnywhere(t, `<?php
trait Registry {
    /** @var list<string> */
    private static $names = [];
    protected static int $count = 0;
    public function read($cls) {
        t('self', self::$names);
        t('static', static::$count);
        t('missing', self::$other);
        t('dynamic', $cls::$names);
        t('parent', parent::$names);
    }
}
`, map[string]string{"self": "string[]", "static": "int", "missing": "?unknown", "dynamic": "?unknown", "parent": "?unknown"})
}
