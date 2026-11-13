package infer_test

import "testing"

// A condition on a property no longer holds past a call that may run
// project code, which may write the property; a builtin call without a
// callable parameter keeps it.
func TestPropertyNarrowingEndsAtProjectCalls(t *testing.T) {
	checkAnywhere(t, `<?php
function refill(): void {}
class Response {
    /** @var string|null */
    private $streamed;
    /** @var string|null */
    public $name;
    /** @var string|null */
    private static $cache;
    public function send(): void {
        ob_start(function (string $buffer) { $this->streamed .= $buffer; return ''; });
    }
    public function content(Response $other) {
        if (!is_null($this->streamed)) {
            strlen('x');
            t('builtin', $this->streamed);
            return $this->streamed;
        }
        t('null', $this->streamed);
        $this->send();
        t('method', $this->streamed);
        if ($other->name === null) {
            refill();
            t('function', $other->name);
        }
        if (self::$cache !== null) {
            new Response();
            t('static', self::$cache);
        }
    }
}
`, map[string]string{
		"builtin": "string", "null": "null",
		"method": "null|string", "function": "null|string", "static": "null|string",
	})
}
