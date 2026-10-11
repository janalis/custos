<?php

interface Marker
{
}

trait Counts
{
    public $count = 0;
}

class Plain implements Marker
{
    use Counts;
}

class Stack extends SplStack
{
}

<warning descr="Choose an explicit comparison for object deduplication.">array_unique([new Plain(), new Plain()])</warning>;
<warning descr="Choose an explicit comparison for object deduplication.">array_unique([new Stack(), new Stack()])</warning>;
<warning descr="Choose an explicit comparison for object deduplication.">array_unique([new DateTime('2020-01-01'), new DateTime('2020-01-01')])</warning>;
<warning descr="Choose an explicit comparison for object deduplication.">array_unique([new ArrayObject([]), new ArrayObject([])], SORT_STRING)</warning>;
