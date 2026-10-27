<?php

class Encoder
{
    // Parameter read only to start the loop, never read after it: nothing is lost.
    public function build(mixed $data): bool
    {
        $ok = true;
        if (is_array($data)) {
            foreach ($data as $key => $data) {
                $ok = $ok && $key !== '' && $data !== null;
            }

            return $ok;
        }

        return false;
    }

    // Parameter known null in the branch, loop key reuses its name; never read after.
    public function all(?string $name = null): iterable
    {
        if (null === $name) {
            foreach (['a' => 1] as $name => $v) {
                yield $name => $v;
            }
        }
    }

    // Branch ends in a throw right after the loop, so the parameter is dead.
    public function send(object $message, array $result): void
    {
        if ([] !== $result) {
            $subject = $message->subject;
            $err = '';
            foreach ($result as $field => $message) {
                $err .= $field . ': ' . $message;
            }
            throw new \RuntimeException($subject . $err);
        }
        echo $message->subject;
    }
}

function stillReported(array $rows): array
{
    <error descr="Loop variable '$rows' overwrites a function parameter.">foreach</error> ($rows as $rows) {
    }

    return $rows; // read after the loop: keep reporting
}

function byReference(array &$rows, array $cells)
{
    <error descr="Loop variable '$rows' overwrites a function parameter.">foreach</error> ($cells as $rows) {}
    <error descr="Loop variable '$cells' overwrites a function parameter.">foreach</error> ($rows as &$cells) {}
}

function named(array $rows, string $n)
{
    <error descr="Loop variable '$rows' overwrites a function parameter.">foreach</error> ($rows as $rows) {}
    echo $$n;
}

function counted(array $rows)
{
    <error descr="Loop variable '$rows' overwrites a function parameter.">foreach</error> ($rows as $rows) {}
    return func_get_args();
}

function looped(array $rows, int $i)
{
    while ($i < count($rows)) {
        <error descr="Loop variable '$rows' overwrites a function parameter.">foreach</error> ($rows as $rows) {}
    }
}

function guarded(array $rows, bool $flag)
{
    try {
        foreach ($rows as $rows) {}
    } finally {
        echo 1;
    }
    switch (1) {
        case 1:
            foreach ([1] as $rows) {}
            break;
    }
    if ($flag) {
        foreach ([2] as $rows) {}
        exit(1);
    }
}

function caught(array $rows)
{
    try {
        <error descr="Loop variable '$rows' overwrites a function parameter.">foreach</error> ($rows as $rows) {}
    } catch (\Exception $e) {
        echo count($rows);
    }
}

function ignored(array $suites, string $name = '')
{
    $out = '';
    <error descr="Loop variable '$name' overwrites a function parameter.">foreach</error> ($suites as $name => $suite) {
        $out .= $name;
    }
    return $out;
}
