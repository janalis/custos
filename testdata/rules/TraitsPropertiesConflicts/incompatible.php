<?php
trait Tracks {
    public $label = 'a';
    protected int $hits = 0;
    public static $registry = [];
    public $loose;
    public ?string $note = null;
    public $ratio = 1;
    public $flag;
}

class Counter {
    use Tracks;
    public <error descr="Counter and trait Tracks both declare property $label.">$label</error> = "b";
    public <error descr="Counter and trait Tracks both declare property $hits.">$hits</error> = 0;
    public <error descr="Counter and trait Tracks both declare property $registry.">$registry</error> = [];
    public <weak_warning descr="Counter and trait Tracks both declare property $loose.">$loose</weak_warning> = null;
    public ?string <weak_warning descr="Counter and trait Tracks both declare property $note.">$note</weak_warning> = NULL;
    public <error descr="Counter and trait Tracks both declare property $ratio.">$ratio</error> = 1.0;
    public $flag = FLAG_DEFAULT;
}

class Visit {
    use Tracks;
    public function __construct(
        public <error descr="Visit and trait Tracks both declare property $label.">$label</error> = 'a',
        protected int <error descr="Visit and trait Tracks both declare property $hits.">$hits</error> = 0,
    ) {}
}

class Typed {
    use Tracks;
    protected ?int <error descr="Typed and trait Tracks both declare property $hits.">$hits</error> = 0;
    public string <error descr="Typed and trait Tracks both declare property $label.">$label</error> = 'a';
}

class BaseCounter {
    public static $label = 'a';
}
class Report extends BaseCounter {
    use <error descr="Report and trait Tracks both declare property $label.">Tracks</error>;
}
