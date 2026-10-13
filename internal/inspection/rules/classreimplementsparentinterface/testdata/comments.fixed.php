<?php
namespace Notes;

interface Marker {}
interface Other {}
interface Unresolved extends Marker, Missing {}
class Base implements Marker, Other, Unresolved {}

// A comment inside the text a fix would delete: reported, left as written.
class BeforeKeyword extends Base /* keep */ implements Marker {}
class BeforeComma extends Base implements Marker /* why */ , \Countable {
    public function count(): int { return 0; }
}
class AfterComma extends Base implements \Countable, /* doc */ Marker {
    public function count(): int { return 0; }
}
class LineComment extends Base implements // first
    Marker // second
{}

// Only the entry whose own deletion is comment-free is fixed.
class Mixed extends Base implements Marker /* first */ {}

// Comments outside the deleted text survive.
class Trailing extends Base /* after */{}
class Kept extends Base implements  /* countable */ \Countable {
    public function count(): int { return 0; }
}
