<?php
class Store
{
    public function fill(&$target) { $target = []; }
}
$s = new Store();
$s->fill(new ArrayObject());
