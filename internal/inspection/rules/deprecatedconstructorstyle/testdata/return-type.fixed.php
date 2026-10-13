<?php
class Query
{
    public function query(): string { return 'select 1'; }
}

class Batch
{
    public function __construct($size) {}
}
