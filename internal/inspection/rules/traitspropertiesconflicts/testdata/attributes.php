<?php
#[Attribute]
class Mapped { public function __construct(...$args) {} }

trait HasParent {
    private $parent;
    private ?int $depth = null;
    protected string $kind = 'node';
}

// Re-declaring a trait property to attach attributes (e.g. ORM mapping) is
// intended: compatible pairs are not reported.
class Folder {
    use HasParent;

    #[Mapped(targetEntity: self::class)]
    private $parent;

    #[Mapped]
    private ?int $depth = null;

    // Incompatible pairs stay errors: PHP refuses to compose the class.
    #[Mapped]
    protected string <error descr="Folder and trait HasParent both declare property $kind.">$kind</error> = 'folder';
}

class Shelf {
    use HasParent;

    private <weak_warning descr="Shelf and trait HasParent both declare property $parent.">$parent</weak_warning>;
}
