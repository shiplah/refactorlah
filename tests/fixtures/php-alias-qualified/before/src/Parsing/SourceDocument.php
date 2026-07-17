<?php

namespace App\Parsing;

use External\Syntax\Expr;
use External\Syntax\Stmt;

final readonly class SourceDocument
{
    public function variable(Stmt\Catch_ $catch): ?Expr\Variable
    {
        return $catch->var instanceof Expr\Variable ? $catch->var : null;
    }
}
