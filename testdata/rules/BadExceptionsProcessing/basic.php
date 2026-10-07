<?php
function load(string $path) {
    <weak_warning descr="Too many statements in this try block; extract some of them so the failing call is obvious.">try</weak_warning> {
        $h = fopen($path, 'r');
        $line = fgets($h);
        fclose($h);
        return trim($line);
    } catch (\ErrorException $problem) {
        error_log($problem->getMessage());
    }

    try {
        $a = 1;
        // comments are not statements
        $b = 2;
        if ($a) { $b++; $b++; $b++; }
    } catch (\LogicException <weak_warning descr="Caught exception is silently discarded; at least log it.">$ignored</weak_warning>) {
        // nothing to do
    } catch (\DomainException | \RangeException <weak_warning descr="Caught exception is dropped; log it or pass it on as the previous exception.">$lost</weak_warning>) {
        throw new \RuntimeException('cannot load');
    } catch (\TypeError $kept) {
        $log = function () use ($kept) { return $kept; };
    } catch (\Error) {
    }
}
