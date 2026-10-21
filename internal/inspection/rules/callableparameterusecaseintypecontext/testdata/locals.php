<?php
class Rebuild {
    protected function rename(\DOMNode &$node, string $name, \DOMDocument $doc): void {
        /** @var \DOMElement $created */
        $created = $node->ownerDocument->createElement($name);
        $created->setAttribute('id', $name);
        $node = $created;
    }

    protected function copy(\DOMNode &$node, \DOMDocument $doc): void {
        $made = $doc->createElement('div');
        $made .= '';
        $node = $made;
    }

    protected function mixedSources(\DOMNode &$node, \DOMDocument $doc, bool $empty): void {
        $made = $doc->createElement('div');
        if ($empty) {
            $made = false;
        }
        $node = <warning descr="Assigning a value of type bool does not match the parameter's declared type.">$made</warning>;
    }

    protected function partlyDocumented(\DOMNode &$node, \DOMDocument $doc, bool $empty): void {
        /** @var \DOMElement $made */
        $made = $doc->createElement('div');
        if ($empty) {
            $made = false;
        }
        $node = <warning descr="Assigning a value of type bool does not match the parameter's declared type.">$made</warning>;
    }

    protected function fromParameter(\DOMNode $node, \DOMElement|false $other): void {
        $node = <warning descr="Assigning a value of type bool does not match the parameter's declared type.">$other</warning>;
    }
}

function endOf(int $end, ?string $date, string $start): array {
    $end = strtotime($start) + 1439 * 60;
    $end = strtotime($start) - true;
    $date = <warning descr="Assigning a value of type int does not match the parameter's declared type.">strtotime($start) * 60</warning>;
    return [$end, $date];
}
