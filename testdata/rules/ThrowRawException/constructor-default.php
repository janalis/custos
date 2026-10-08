<?php
class GoneException extends \RuntimeException
{
    public function __construct(string $message = 'Resource is gone.', int $code = 0, ?\Throwable $previous = null)
    {
        parent::__construct($message, $code, $previous);
    }
}

class EmptyDefaultException extends \RuntimeException
{
    public function __construct(string $message = '', int $code = 0, ?\Throwable $previous = null)
    {
        parent::__construct($message, $code, $previous);
    }
}

function fail(int $n): void
{
    if ($n === 1) {
        throw new GoneException();
    }
    throw <weak_warning descr="Pass a message when throwing this exception.">new EmptyDefaultException()</weak_warning>;
}
