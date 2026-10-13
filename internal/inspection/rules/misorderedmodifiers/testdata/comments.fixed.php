<?php

abstract class Template
{
    public /* not final on purpose */ static function make() { return 'base'; }
    abstract /* static later */ protected function keep();
    final # private
        public static function sealed() {}

    public /* was protected */ static function open() { return 'open'; }
    final // abstract one day
        protected /* final */ static function guard() {}
    final /** @deprecated */ public function done() {}
}
