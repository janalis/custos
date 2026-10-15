<?php
final class SlotStore
{
    public function &replace(&$slot) { $slot = 0; return $slot; }
    public function read() { return []; }
    public static function token() { return 'token'; }
}
function replaceStore(&$store) { $store = false; }
function acceptSlot(&$slot) {}
function makeSlot() { return []; }

function unavailableReceivers()
{
    $store = new SlotStore();
    $store->replace($store);
    $store->replace(makeSlot());
    $reader = new SlotStore();
    replaceStore($reader);
    acceptSlot($reader->read());
}

function restoredReceiver()
{
    $store = new SlotStore();
    $store->replace($store);
    $store = new SlotStore();
    $store->replace(<warning descr="Pass a variable here: this parameter is taken by reference.">makeSlot()</warning>);
    $store->replace(<warning descr="Pass a variable here: this parameter is taken by reference.">$store->read()</warning>);
    $store->replace(<warning descr="Pass a variable here: this parameter is taken by reference.">SlotStore::token()</warning>);
    $store->replace(<warning descr="Pass a variable here: this parameter is taken by reference.">new stdClass()</warning>);
    acceptSlot(<warning descr="Pass a variable here: this parameter is taken by reference.">$store->read()</warning>);
}
