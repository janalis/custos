<?php
function phpFiles(string $root): array
{
    $out = [];
    $it = new RecursiveIteratorIterator(new RecursiveDirectoryIterator($root, FilesystemIterator::SKIP_DOTS));
    foreach ($it as $file) {
        if (!$file->isFile() || 'php' !== $file->getExtension()) {
            continue;
        }
        $out[] = $file->getPathname();
    }
    foreach ((new RecursiveIteratorIterator(new RecursiveDirectoryIterator($root))) as $f) {
        if ($f->getExtension() === 'css') {
            $out[] = $f;
        }
    }
    return $out;
}

function cssFiles(string $root): RegexIterator
{
    return new RegexIterator(new RecursiveIteratorIterator(new RecursiveDirectoryIterator($root)), '/\.css$/');
}

function prune(string $root): void
{
    $it = <warning descr="Include parent nodes when traversing directories.">new RecursiveIteratorIterator(new RecursiveDirectoryIterator($root))</warning>;
    foreach ($it as $node) {
        $node->isDir() ? rmdir($node->getPathname()) : unlink($node->getPathname());
    }
}

function shared(string $root): iterable
{
    $it = <warning descr="Include parent nodes when traversing directories.">new RecursiveIteratorIterator(new RecursiveDirectoryIterator($root))</warning>;
    foreach ($it as $file) {
        if ($file->isFile()) {
            echo $file;
        }
    }
    return $it;
}

function twice(string $root): void
{
    $it = <warning descr="Include parent nodes when traversing directories.">new RecursiveIteratorIterator(new RecursiveDirectoryIterator($root))</warning>;
    foreach ($it as $file) {
        if ($file->isFile()) {
            echo $file;
        }
    }
    foreach ($it as [$a, $b]) {
        echo $a;
    }
    $other = <warning descr="Include parent nodes when traversing directories.">new RecursiveIteratorIterator(new RecursiveDirectoryIterator($root))</warning>;
    $keep = fn () => $other;
    $cb = function () use ($root) {
        foreach ($it as $file) {
            $file->isFile();
        }
    };
    $list = <warning descr="Include parent nodes when traversing directories.">new RecursiveIteratorIterator(new RecursiveDirectoryIterator($root))</warning>;
    foreach ($list as $entry) {
        $entry->getFilename();
        $entry->{$m}();
        $entry->isFile;
    }
    $items[0] = <warning descr="Include parent nodes when traversing directories.">new RecursiveIteratorIterator(new RecursiveDirectoryIterator($root))</warning>;
    $w = new IteratorIterator(<warning descr="Include parent nodes when traversing directories.">new RecursiveIteratorIterator(new RecursiveDirectoryIterator($root))</warning>);
    $x = new RegexIterator($it, <warning descr="Include parent nodes when traversing directories.">new RecursiveIteratorIterator(new RecursiveDirectoryIterator($root))</warning>);
    run(<warning descr="Include parent nodes when traversing directories.">new RecursiveIteratorIterator(new RecursiveDirectoryIterator($root))</warning>);
    $y = fn () => <warning descr="Include parent nodes when traversing directories.">new RecursiveIteratorIterator(new RecursiveDirectoryIterator($root))</warning>;
}

$top = new RecursiveIteratorIterator(new RecursiveDirectoryIterator('.'));
foreach ($top as $file) {
    if ($file->isFile()) {
        echo $file;
    }
}
