<?php
interface Event { public function getParam(string $name); }

function annotate(Event $e): void
{
    $spec = $e->getParam('inputSpec');
    $spec['allow_empty'] = true;            // may be an ArrayObject
}

function castLocal(Event $e): void
{
    $spec = (array) $e->getParam('inputSpec');
    <weak_warning descr="Value is only written here and never read; the write is lost.">$spec['a']</weak_warning> = 1;
}

function mixedWrites(Event $e): void
{
    $spec = $e->getParam('inputSpec');
    <weak_warning descr="Value is only written here and never read; the write is lost.">$spec['a']</weak_warning> = 1;
    <weak_warning descr="Value is only written here and never read; the write is lost.">$spec++</weak_warning>;
}

function knownArray(): void
{
    $spec = [];
    <weak_warning descr="Value is only written here and never read; the write is lost.">$spec['allow_empty']</weak_warning> = true;
}
