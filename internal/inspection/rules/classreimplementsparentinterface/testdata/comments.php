<?php
namespace Notes;

interface Marker {}
interface Other {}
interface Unresolved extends Marker, Missing {}
class Base implements Marker, Other, Unresolved {}

// A comment inside the text a fix would delete: reported, left as written.
class BeforeKeyword extends Base /* keep */ implements <warning descr="'\Notes\Marker' is already implemented by '\Notes\Base'; remove it here.">Marker</warning> {}
class BeforeComma extends Base implements <warning descr="'\Notes\Marker' is already implemented by '\Notes\Base'; remove it here.">Marker</warning> /* why */ , \Countable {
    public function count(): int { return 0; }
}
class AfterComma extends Base implements \Countable, /* doc */ <warning descr="'\Notes\Marker' is already implemented by '\Notes\Base'; remove it here.">Marker</warning> {
    public function count(): int { return 0; }
}
class LineComment extends Base implements // first
    <warning descr="'\Notes\Marker' is already implemented by '\Notes\Base'; remove it here.">Marker</warning> // second
{}

// Only the entry whose own deletion is comment-free is fixed.
class Mixed extends Base implements <warning descr="'\Notes\Marker' is already implemented by '\Notes\Base'; remove it here.">Marker</warning> /* first */, <warning descr="'\Notes\Other' is already implemented by '\Notes\Base'; remove it here.">Other</warning> {}

// Comments outside the deleted text survive.
class Trailing extends Base implements <warning descr="'\Notes\Marker' is already implemented by '\Notes\Base'; remove it here.">Marker</warning>/* after */{}
class Kept extends Base implements <warning descr="'\Notes\Marker' is already implemented by '\Notes\Base'; remove it here.">Marker</warning>, /* countable */ \Countable {
    public function count(): int { return 0; }
}
