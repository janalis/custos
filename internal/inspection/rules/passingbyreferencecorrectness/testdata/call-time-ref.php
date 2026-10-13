<?php
class Store
{
    public function fill(&$target) { $target = array(); }
    public function plain() { return array(); }
}
$s = new Store();
$s->fill(&$s->plain());
