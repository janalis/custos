<?php
abstract class Node {
    /** @return static */
    abstract public function copy(): static;
}
