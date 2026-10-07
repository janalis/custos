<?php
trait Tracks { private $trail; }

class Model {
    use Tracks;
    private $secret;
    protected $table;
    protected $visible;
    public $id;
    protected static $registry;
    const KIND = 'model';
}
class Record extends Model {
    protected $dirty;
    protected int $limit = 10;
    protected array $columns = ['a', 'b'];
}

/** @property string $virtual */
class Invoice extends Record {
    protected $table = 'invoices'; // overrides the default: kept
    protected <weak_warning descr="Property 'dirty' is already declared in \Record; drop this re-declaration.">$dirty</weak_warning>;
    public <weak_warning descr="Property 'id' is already declared in \Model; drop this re-declaration.">$id</weak_warning> = NULL;
    public <weak_warning descr="\Model already has a private property with this name; consider a different name.">$secret</weak_warning>;
    private <weak_warning descr="\Tracks already has a private property with this name; consider a different name.">$trail</weak_warning>;
    private <weak_warning descr="Property 'visible' is already declared in \Model; drop this re-declaration.">$visible</weak_warning>, $fresh;
    /** @var string */
    protected $virtual;
    protected int <weak_warning descr="Property 'limit' is already declared in \Record; drop this re-declaration.">$limit</weak_warning> = 10;
    protected array <weak_warning descr="Property 'columns' is already declared in \Record; drop this re-declaration.">$columns</weak_warning> = [ 'a', 'b' ];
}
