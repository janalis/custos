<?php
function evalReadsImport(int $id)
{
    $run = function () use ($id) {
        return eval('return $id + 1;');
    };
    return $run();
}

function evalReadsParameter(array $rows)
{
    $rows[] = 1;
    return eval('return count($rows);');
}

function argsReturned(string $label, $count): array
{
    $label .= '!';
    $count++;
    return func_get_args();
}

function oneArgReturned(string $label)
{
    $label .= '!';
    return func_get_arg(0);
}

function argsOfAnArrowFunction(string $label)
{
    <weak_warning descr="Value is only written here and never read; the write is lost.">$label</weak_warning> .= '!';
    return fn () => func_get_args();
}

function argsDoNotReadLocals()
{
    $buffer = [];
    <weak_warning descr="Value is only written here and never read; the write is lost.">$buffer[]</weak_warning> = 1;
    return func_get_args();
}
