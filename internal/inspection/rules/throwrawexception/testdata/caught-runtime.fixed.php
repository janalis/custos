<?php

namespace App;

use RuntimeException as Retryable;

function attempt(callable $job): string
{
    try {
        if (!$job()) {
            throw new \Exception('failed');
        }
    } catch (\LogicException | Retryable $e) {
        return 'retry';
    } catch (\Exception $e) {
        return 'fail';
    }
    try {
        throw new \RuntimeException('outer');
    } catch (\Exception $e) {
        return 'fail';
    }
    try {
        $job();
    } catch (Retryable $e) {
        throw new \RuntimeException('again');
    }
    return 'ok';
}
