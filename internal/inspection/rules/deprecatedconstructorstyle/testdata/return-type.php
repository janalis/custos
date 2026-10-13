<?php
class Query
{
    public function <error descr="Class 'Query' uses an old-style constructor; rename it to __construct.">query</error>(): string { return 'select 1'; }
}

class Batch
{
    public function <error descr="Class 'Batch' uses an old-style constructor; rename it to __construct.">Batch</error>($size) {}
}
