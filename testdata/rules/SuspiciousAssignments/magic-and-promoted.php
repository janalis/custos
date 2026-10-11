<?php
// Writes that run code (offsetSet, __set), promoted parameters, and calls
// between two fall-through writes of shared state.

final class Journal implements ArrayAccess
{
    public array $lines = [];
    public string $title = '';

    public function offsetSet(mixed $key, mixed $value): void { $this->lines[] = "$key=$value"; }
    public function offsetGet(mixed $key): mixed { return null; }
    public function offsetExists(mixed $key): bool { return false; }
    public function offsetUnset(mixed $key): void {}
    public function __set(string $name, mixed $value): void { $this->lines[] = "$name:=$value"; }

    public function stamp(string $kind, bool $late): void
    {
        $this->phase = 'begin';
        $this->phase = 'end';
        $this->title = 'draft';
        <error descr="$this->title is overwritten right after being assigned.">$this->title</error> = 'final';
        if ($late) {
            $this->delay = 1;
        }
        $this->delay = 0;
        switch ($kind) {
            case 'open':
                $this->marker = 'o';
                $this['slot'] = 'o';
                $this->title = 'o';
            case 'shut':
                $this->marker = 's';
                $this['slot'] = 's';
                [$this->marker, $rest] = ['t', 'u'];
                <error descr="This write overwrites a value set in a previous case; a 'break' may be missing.">$this->title</error> = 's';
                break;
        }
    }
}

function record_events(): array
{
    $journal = new Journal();
    $journal['event'] = 'start';
    $journal['event'] = 'stop';
    $journal->state = 'open';
    $journal->state = 'closed';
    return $journal->lines;
}

function plain_containers(): array
{
    $bag = new ArrayObject();
    $bag['size'] = 1;
    <error descr="$bag['size'] is overwritten right after being assigned.">$bag['size']</error> = 2;
    $box = new stdClass();
    $box->size = 1;
    <error descr="$box->size is overwritten right after being assigned.">$box->size</error> = 2;
    $name = 'size';
    $box->$name = 3;
    <error descr="$box->$name is overwritten right after being assigned.">$box->$name</error> = 4;
    return [$bag, $box];
}

final class Shift
{
    public array $tags;

    public function __construct(public string $name, private readonly int $rank, array $tags = [], $spare = null)
    {
        $name = strtoupper('fallback');
        $rank = 0;
        <error descr="Parameter is overwritten before its value is used.">$spare</error> = 'unused';
        $this->tags = [$name, $rank, $spare, ...$tags];
    }
}

final class Panel
{
    public string $mode = '';
    public array $out = [];

    public function render(): void { $this->out[] = $this->mode; }

    public function show(string $kind): array
    {
        switch ($kind) {
            case 'both':
                $this->mode = 'header';
                $label = 'both';
                $this->render();
            case 'body':
                $this->mode = 'body';
                <error descr="This write overwrites a value set in a previous case; a 'break' may be missing.">$label</error> = 'body';
                $this->render();
                break;
            case 'plain':
                $this->mode = 'plain';
            case 'bare':
                <error descr="This write overwrites a value set in a previous case; a 'break' may be missing.">$this->mode</error> = 'bare';
                break;
            case 'late':
                $this->mode = 'late';
            case 'later':
                $this->mode = $this->pick();
                break;
        }
        return [$this->out, $label ?? ''];
    }

    private function pick(): string { return $this->mode . '!'; }
}
