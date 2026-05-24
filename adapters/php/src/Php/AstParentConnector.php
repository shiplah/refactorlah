<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use PhpParser\Node;

use function is_array;

final class AstParentConnector
{
    /** @param list<Node> $ast */
    public function attach(array $ast): void
    {
        foreach ($ast as $node) {
            $this->attachParent($node, null);
        }
    }

    private function attachParent(Node $node, ?Node $parent): void
    {
        $node->setAttribute('parent', $parent);
        foreach ($node->getSubNodeNames() as $name) {
            $child = $node->$name;
            if ($child instanceof Node) {
                $this->attachParent($child, $node);
                continue;
            }

            if (!is_array($child)) {
                continue;
            }

            foreach ($child as $nested) {
                if ($nested instanceof Node) {
                    $this->attachParent($nested, $node);
                }
            }
        }
    }
}
