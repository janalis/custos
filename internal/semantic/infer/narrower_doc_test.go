package infer_test

import "testing"

// A member's doc type wins over its declaration when it names subtypes of
// what the declaration names; an unrelated class, a wider doc or a doc
// adding null does not.
func TestNarrowerMemberDoc(t *testing.T) {
	checkAnywhere(t, `<?php
interface Message {}
class Text implements Message {}
class Html extends Text {}
class Other {}
class C {
    public function __construct(
        /** @var Text */ private Message $message,
        /** @var Html|Text|null */ private ?Message $maybe,
        /** @var Other */ private Message $unrelated,
        /** @var Text|null */ private Message $widened,
        /** @var Message */ private Message $same,
    ) {}
    /** @return Html */
    public function make(): Text { return new Html(); }
    public function f() {
        t('narrowed', $this->message);
        t('nullable', $this->maybe);
        t('unrelated', $this->unrelated);
        t('widened', $this->widened);
        t('same', $this->same);
        t('return', $this->make());
    }
}
`, map[string]string{
		"narrowed": `\Text`, "nullable": `\Html|\Text|null`, "unrelated": `\Message`, "widened": `\Message`, "same": `\Message`, "return": `\Html`,
	})
}

// A parameter documented as a subtype of its declaration takes the doc type.
func TestNarrowerParamDoc(t *testing.T) {
	checkAnywhere(t, `<?php
interface Message {}
class Text implements Message {}
class Other {}
/**
 * @param Text $narrowed
 * @param Text|null $nullable
 * @param Other $unrelated
 */
function f(Message $narrowed, ?Message $nullable, Message $unrelated) {
    t('narrowed', $narrowed);
    t('nullable', $nullable);
    t('unrelated', $unrelated);
}
`, map[string]string{"narrowed": `\Text`, "nullable": `\Text|null`, "unrelated": `\Message`})
}
