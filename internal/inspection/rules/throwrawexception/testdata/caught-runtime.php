<?php

namespace App;

use RuntimeException as Retryable;

function attempt(callable $job): string
{
    try {
        if (!$job()) {
            throw new <weak_warning descr="Throw a more specific exception class than \Exception.">\Exception</weak_warning>('failed');
        }
    } catch (\LogicException | Retryable $e) {
        return 'retry';
    } catch (\Exception $e) {
        return 'fail';
    }
    try {
        throw new <weak_warning descr="Throw a more specific exception class than \Exception.">\Exception</weak_warning>('outer');
    } catch (\Exception $e) {
        return 'fail';
    }
    try {
        $job();
    } catch (Retryable $e) {
        throw new <weak_warning descr="Throw a more specific exception class than \Exception.">\Exception</weak_warning>('again');
    }
    return 'ok';
}
