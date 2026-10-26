<?php
namespace Psr\Log {
    interface LoggerInterface {}
    abstract class AbstractLogger implements LoggerInterface {}
}

namespace App {
    use Psr\Log\AbstractLogger;

    final class StderrLogger extends AbstractLogger
    {
        public function log($level, $message): void
        {
            error_log($level . ': ' . $message);
            <error descr="Debug output call; remove it if it was left over from debugging.">print_r($message)</error>;
        }
    }

    final class Mailer
    {
        public function send($message): void
        {
            <error descr="Debug output call; remove it if it was left over from debugging.">error_log($message)</error>;
        }
    }

    function report($message): void
    {
        <error descr="Debug output call; remove it if it was left over from debugging.">error_log($message)</error>;
    }
}
