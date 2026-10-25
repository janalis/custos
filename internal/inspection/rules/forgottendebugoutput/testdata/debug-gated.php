<?php

class Mailer
{
    public int $debugLevel = 0;
    public string $debugOutput = 'echo';
    private bool $verbose = false;

    protected function edebug(string $str): void
    {
        switch ($this->debugOutput) {
            case 'error_log':
                error_log($str);
                break;
            default:
                echo $str;
        }
    }

    public function dumpErrors(array $errors): void
    {
        if ($this->debugLevel >= 1) {
            foreach ($errors as $e) {
                print_r($e);
            }
        } elseif (self::DEBUG_SERVER > 2) {
            var_dump($errors);
        } else {
            <error descr="Debug output call; remove it if it was left over from debugging.">var_dump($errors)</error>;
        }
        if ($this->verbose) {
            <error descr="Debug output call; remove it if it was left over from debugging.">var_dump($errors)</error>;
        }
        if (<error descr="Debug output call; remove it if it was left over from debugging.">var_dump($errors)</error> && isDebug()) {
        }
        switch (<error descr="Debug output call; remove it if it was left over from debugging.">print_r($errors)</error>) {
            case DEBUG:
        }
        if (isDebug()) {
            $f = function () use ($errors) {
                <error descr="Debug output call; remove it if it was left over from debugging.">var_dump($errors)</error>;
            };
        }
    }
}
