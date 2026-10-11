<?php

abstract class Template
{
    public /* not final on purpose */ static function make() { return 'base'; }
    abstract /* static later */ protected function keep();
    final # private
        public static function sealed() {}

    <weak_warning descr="Reorder modifiers as: public static.">static /* was protected */ public</weak_warning> function open() { return 'open'; }
    <weak_warning descr="Reorder modifiers as: final protected static.">static // abstract one day
        protected /* final */ final</weak_warning> function guard() {}
    <weak_warning descr="Reorder modifiers as: final public.">Public /** @deprecated */ FINAL</weak_warning> function done() {}
}
