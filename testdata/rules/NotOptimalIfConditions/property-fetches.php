<?php
function load(string $n) { return $n; }

final class Folder {
    public bool $open = false;
    public array $files = [];
    public bool $isEmpty { get => $this->files === []; }
    public string $name { set => strtolower($value); }
    public string $title { set { $this->title = trim($value); } }
    public string $alias { set { load($value); } }
    public function __construct(public ?string $label = null, public int $size = 0 { get => $this->size * 2; }) {}
}
final class Lazy {
    public function __get(string $n) { return load($n); }
}
/** @property string $magic */
final class Documented {}
interface HasTitle {
    public string $title { get; }
}
abstract class Base {
    abstract public string $code { get; }
    public int $count = 0;
}
final class Page {
    public int $id = 0;
    public bool $hasContent {
        get {
            return load('sections') !== '';
        }
    }
}

function scan(Folder $f, Lazy $l, $any, string $p, Documented $d, HasTitle $h, Base $b, ?Folder $nf, Folder|Lazy $u, \stdClass $s, int $n, string $k) {
    if (strlen($p) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$f->open</weak_warning>) {}
    if (trim($p) || <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$f->label</weak_warning>) {}
    if (strlen($p) > 3 && !$f->isEmpty) {}      // get hook: computed, impure
    if (is_dir($p) && $l->cached) {}            // __get(): computed, impure
    if (strlen($p) > 3 && $any->flag) {}        // untyped receiver: unknown
    if ($f->isEmpty || $f->open) {}             // computed neighbour: impure
    if (strlen($p) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$f->name</weak_warning>) {} // short set hook: backed
    if (strlen($p) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$f->title</weak_warning>) {} // set hook writes the backing store
    if (strlen($p) > 3 && $f->alias) {}         // virtual
    if (strlen($p) > 3 && $f->size) {}          // promoted with a get hook
    if (strlen($p) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$d->magic</weak_warning>) {} // @property without __get(): stored
    if (strlen($p) > 3 && $h->title) {}         // interface property
    if (strlen($p) > 3 && $b->code) {}          // abstract property
    if (strlen($p) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$b->count</weak_warning>) {}
    if (strlen($p) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$nf?->open</weak_warning>) {}
    if (strlen($p) > 3 && $u->open) {}          // one class has __get()
    if (strlen($p) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$s->anything</weak_warning>) {} // dynamic property
    if (strlen($p) > 3 && $n->prop) {}          // not an object: unknown
    if (strlen($p) > 3 && $f->{$k}) {}          // dynamic name: unknown
    if (strlen($p) > 3 && $f->$k) {}            // dynamic name: unknown
    if (strlen($p) > 3 && $missing->prop) {}    // unresolved receiver
    if ((fn() => $any->flag) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$f->open</weak_warning>) {} // closure bodies do not count
}

function real(iterable $pages) {
    foreach ($pages as $page) {
        if (!$page instanceof Page) {
            continue;
        }
        $slug = (string) $page->id;
        if ('' === $slug || filter_var($slug, \FILTER_VALIDATE_URL) || !$page->hasContent) {
            continue;
        }
    }
}
