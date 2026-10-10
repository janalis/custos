<?php
interface Starter { public function start(); } class Worker implements Starter { #[\Override] public function start() {} }
