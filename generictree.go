/*
<!--
Copyright (c) 2021 Christoph Berger. Some rights reserved.
Use of this text is governed by a Creative Commons Attribution Non-Commercial
Share-Alike License that can be found in the LICENSE.txt file.

The source code contained in this file may import third-party source code
whose licenses are provided in the respective license files.
-->

<!--
NOTE: The comments in this file are NOT godoc compliant. This is not an oversight.

Comments and code in this file are used for describing and explaining a particular topic to the reader. While this file is a syntactically valid Go source file, its main purpose is to get converted into a blog article. The comments were created for learning and not for code documentation.
-->

+++
title = "How I turned a binary search tree into a generic data structure with go2go"
description = "Steps taken to turn a binary search tree that has integer keys and string data into a generic tree that can have arbitrary (sortable) key types and arbitrary payload types, thanks to the upcoming generics feature in Go"
author = "Christoph Berger"
email = "chris@appliedgo.net"
date = "2021-07-07"
draft = false
categories = ["Algorithms And Data Structures"]
tags = ["Tree", "Balanced Tree", "Binary Tree", "generics"]
articletypes = ["Tutorial"]
+++

Some time ago I wrote about how to create a balanced binary search tree. The search keys and the data payload were both plain strings. Now it is time to get rid of this limitation. go2go lets us do that while waiting for the official generics release.

<!--more-->

___

**Update:** Go type parameters have changed since `go2go`. The article has been updated to match the syntax and semantics of type parameters in Go 1.18 and use the `cmp` package of Go 1.21 instead of `constraints`.
___

Warning: This article is super boring! It turned out that converting a container type into a generic container type is quite straightforward with `go2go` and shows no surprises.

Which is actually a good sign.

It is a good sign because adding generic data types and functions to a programming language is dead easy... to get wrong. Hence the Go team went to great lengths, and took all possible precautions, to design generics that don't suck. And IMHO, the current [proposal](https://blog.golang.org/generics-proposal) should appeal even to the ones who were skeptical about adding generics to Go *at all*.

With the current generics design, it would seem fairly easy to create new generic data structures and generic functions, but what about sifting through old code to make it generic? Will there be any footguns?

Let's find out.

![Generic Trees](generictree.jpg)

## The *status quo* of the search tree code

In [this article](https://appliedgo.net/bintree), I created a binary tree, and in [another article](https://appliedgo.net/balancedtree), I turned the tree into a balanced tree (with AVL balancing logic). Both the search key and the payload data are of type `string`.

```go
type Node struct {
	Value  string
	Data   string
	Left   *Node
	Right  *Node
	height int
}
```

## What to change

Obviously, I need to change the types of the fields `Value` and `Data`.

Then, all functions that take or return either of these two fields, or that take a `Node` and access the fields through the `Node` struct, need to be adjusted. This applies to functions like `Insert()` or `min()`, for example.


Let's walk through the code and adjust it as required.

*/

// As always, the code starts with package and import statements, as the whole blog article is generated from a single, compilable Go source file.
//
// Note the import of the 'cmp' package (added in Go 1.21). This package provides types and functions for comparing ordered values, including the `Ordered` constraint that I need for being able to compare and sort the nodes.
package main

import (
	"cmp"
	"fmt"
	"strings"
)

/*

### Step 1: Change existing types

First, I take the `Node` struct shown above, and change the `Value` and `Data` fields
from `string` to the new generic `Value` and `Data` types. While the Value type must be ordered, the Data type can be anything.

This turns the Node struct itself into a generic type that I now must declare with
appropriate type parameters. In general, any generic types declared inside a struct bubble up to the struct type declaration.

Note that the `*Node` pointer types inside the struct also need to be properly parameterized.


*/
// type Node struct {\
//    Value string\
//    Data string\
//    Right *Node\
//    Left *Node
type Node[Value cmp.Ordered, Data any] struct {
	Value  Value
	Data   Data
	Left   *Node[Value, Data]
	Right  *Node[Value, Data]
	height int
}

/*

*(In the comment block, this is how the struct looked before.)*

When instantiating a `Node`, concrete types for the Value and Data parameters must be supplied.
Then the fields `Value` and `Data` get instantiated to the given concrete types.

Example: `n := *Node{uint16, []byte}`



### Step 2: change functions and methods

Now let's look through all the functions and methods and make them polymorphic.

Wherever a function receives a `Node` value, or a value string or data string,
I need to change this to the respective generic type, for example, `Node[Value, Data]`.

The same applies to method receivers.

*/
// Here, you can see why I need an `Ordered` constraint.
// Type `T` must support comparison operations, otherwise `a > b` would fail
// at runtime if T is instantiated with a non-comparable type.
func max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Besides the receiver type, nothing needs to be changed here.
// `*Node` becomes `*Node[Value, Data]`.\
// Later, when instantiating a struct of type `Node`, concrete types
// need to be supplied for `Value` and `Data`.
func (n *Node[Value, Data]) Height() int {
	if n == nil {
		return 0
	}
	return n.height
}

func (n *Node[Value, Data]) Bal() int {
	return n.Right.Height() - n.Left.Height()
}

// Here is the first occurrence of generic parameters and return types.\
// `value, data string` is now \
// `value Value, data Data`.\
// The function body remains untouched, as all operations on `value`, `data`, `n.Value`, or `n.Data`
// work the same, even though the concrete types for `Value` and `Data` are not known yet.\
// Especially, `==` and `<` work fine for the `Value` type because of the `Ordered` type constraint.
func (n *Node[Value, Data]) Insert(value Value, data Data) *Node[Value, Data] {
	if n == nil {
		return &Node[Value, Data]{
			Value:  value,
			Data:   data,
			height: 1,
		}
	}
	if n.Value == value {
		n.Data = data
		return n
	}

	if value < n.Value {
		n.Left = n.Left.Insert(value, data)
	} else {
		n.Right = n.Right.Insert(value, data)
	}

	n.height = max(n.Left.Height(), n.Right.Height()) + 1

	return n.rebalance()
}

// From here onwards, the same pattern repeats. The function signatures receive generic parameters for the Node type, and the function bodies remain largely unmodified. \
// `#boring`
func (n *Node[Value, Data]) rotateLeft() *Node[Value, Data] {
	r := n.Right
	n.Right = r.Left
	r.Left = n
	n.height = max(n.Left.Height(), n.Right.Height()) + 1
	r.height = max(r.Left.Height(), r.Right.Height()) + 1
	return r
}

func (n *Node[Value, Data]) rotateRight() *Node[Value, Data] {
	l := n.Left
	n.Left = l.Right
	l.Right = n
	n.height = max(n.Left.Height(), n.Right.Height()) + 1
	l.height = max(l.Left.Height(), l.Right.Height()) + 1
	return l
}

func (n *Node[Value, Data]) rotateRightLeft() *Node[Value, Data] {
	n.Right = n.Right.rotateRight()
	n = n.rotateLeft()
	n.height = max(n.Left.Height(), n.Right.Height()) + 1
	return n
}

func (n *Node[Value, Data]) rotateLeftRight() *Node[Value, Data] {
	n.Left = n.Left.rotateLeft()
	n = n.rotateRight()
	n.height = max(n.Left.Height(), n.Right.Height()) + 1
	return n
}

func (n *Node[Value, Data]) rebalance() *Node[Value, Data] {
	switch {
	case n.Bal() < -1 && n.Left.Bal() == -1:
		return n.rotateRight()
	case n.Bal() > 1 && n.Right.Bal() == 1:
		return n.rotateLeft()
	case n.Bal() < -1 && n.Left.Bal() == 1:
		return n.rotateLeftRight()
	case n.Bal() > 1 && n.Right.Bal() == -1:
		return n.rotateRightLeft()
	}
	return n
}

func (n *Node[Value, Data]) Find(s Value) (Data, bool) {

	if n == nil {
		// Interesting detail: `go2go` has no dedicated expression for "zero value of type T" (yet).
		// This is resolved here by instantiating a variable of type T and returning that variable.\
		// An alternate way is shown below, and a third alternative is to use named return parameters
		// and use a naked `return` statement.
		var zero Data
		return zero, false
	}

	switch {
	case s == n.Value:
		return n.Data, true
	case s < n.Value:
		return n.Left.Find(s)
	default:
		return n.Right.Find(s)
	}
}

func (n *Node[Value, Data]) Dump(i int, lr string) {
	if n == nil {
		return
	}
	indent := ""
	if i > 0 {
		indent = strings.Repeat(" ", (i-1)*4) + "+" + lr + "--"
	}
	fmt.Printf("%s%v[%d,%d]\n", indent, n.Value, n.Bal(), n.Height())
	n.Left.Dump(i+1, "L")
	n.Right.Dump(i+1, "R")
}

type Tree[Value cmp.Ordered, Data any] struct {
	Root *Node[Value, Data]
}

func (t *Tree[Value, Data]) Insert(value Value, data Data) {
	t.Root = t.Root.Insert(value, data)
	if t.Root.Bal() < -1 || t.Root.Bal() > 1 {
		t.rebalance()
	}
}

func (t *Tree[Value, Data]) rebalance() {
	if t == nil || t.Root == nil {
		return
	}
	t.Root = t.Root.rebalance()
}

func (t *Tree[Value, Data]) Find(s Value) (Data, bool) {
	if t == nil || t.Root == nil {
		// Same situation as in method `Find` above.\
		// Here, we use `new` to create a zero value on the fly.\
		// `new` returns a pointer, and hence we need to add the dereferencing operator.
		return *new(Data), false
	}
	return t.Root.Find(s)
}

func (t *Tree[Value, Data]) Traverse(n *Node[Value, Data], f func(*Node[Value, Data])) {
	if n == nil {
		return
	}
	t.Traverse(n.Left, f)
	f(n)
	t.Traverse(n.Right, f)
}

func (t *Tree[Value, Data]) PrettyPrint() {

	printNode := func(n *Node[Value, Data], depth int) {
		fmt.Printf("%s%v\n", strings.Repeat("  ", depth), n.Value)
	}
	var walk func(*Node[Value, Data], int)
	walk = func(n *Node[Value, Data], depth int) {
		if n == nil {
			return
		}
		walk(n.Right, depth+1)
		printNode(n, depth)
		walk(n.Left, depth+1)
	}

	walk(t.Root, 0)
}

func (t *Tree[Value, Data]) Dump() {
	t.Root.Dump(0, "")
}

/*
## How to use the new generic tree type

Now is the moment where I can instantiate the generic `Tree[Value, Data]` type into something tangible like `Tree[int,string]`.

*/

func main() {
	values := []string{"d", "b", "g", "g", "c", "e", "a", "h", "f", "i", "j", "l", "k"}
	data := []string{"delta", "bravo", "golang", "golf", "charlie", "echo", "alpha", "hotel", "foxtrot", "india", "juliett", "lima", "kilo"}

	// Here, Tree gets instantiated with the `string` type for both Value and Data.
	// This is basically the same tree as in the original article about balanced trees.
	tree := &Tree[string, string]{}
	for i := 0; i < len(values); i++ {
		tree.Insert(values[i], data[i])
	}

	fmt.Print("\n*** Tree with string search values and string data ***\n\n")
	fmt.Print("Sorted values: | ")
	// As with `*Tree` above, `*Node` also needs to get instantiated with concrete types.
	tree.Traverse(tree.Root, func(n *Node[string, string]) { fmt.Print(n.Value, ": ", n.Data, " | ") })
	fmt.Println()

	fmt.Println("Pretty print (turned 90° anti-clockwise):")
	tree.PrettyPrint()

	// Let's try the same with integers as search values.
	keys := []int{4, 2, 7, 7, 3, 5, 1, 8, 6, 9, 10, 12, 11}
	// No new `data` slice here. It remains the same slice of strings.

	// This time, Tree gets instantiated with `int` and `string` for Value and Data, respectively.
	intTree := &Tree[int, string]{}
	for i := 0; i < len(keys); i++ {
		intTree.Insert(keys[i], data[i])
	}

	fmt.Print("\n*** Tree with int search values and string data ***\n\n")
	fmt.Print("Sorted values: | ")
	intTree.Traverse(intTree.Root, func(n *Node[int, string]) { fmt.Print(n.Value, ": ", n.Data, " | ") })
	fmt.Println()

	fmt.Println("Pretty print")
	intTree.PrettyPrint()

	/*
		### How about creating a search tree of search trees?

		Let's feed a search tree with search trees as payload data. \
		Because why not?\
		And because doing this can answer an interesting question: Will the syntax of nested generic type instantiatons become unwieldy?
	*/
	// The search values shall be integers.
	keys = []int{3, 1, 2}
	// I am lazy here and use the existing "string, string" tree thrice.
	trees := []*Tree[string, string]{tree, tree, tree}

	// This is a nested instantiation of generic types. Nice detail: the syntax really remains readable.
	treeTree := &Tree[int, *Tree[string, string]]{}
	for i := 0; i < len(keys); i++ {
		treeTree.Insert(keys[i], trees[i])
	}

	fmt.Print("\n*** Tree with int search values and Tree[string, string] data ***\n\n")
	fmt.Print("Sorted values: | ")
	// As with `*Tree` above, `*Node` also needs to get instantiated with concrete types.
	treeTree.Traverse(treeTree.Root, func(n *Node[int, *Tree[string, string]]) { fmt.Print(n.Value, ": ", n.Data, " | ") })
	fmt.Println()

	fmt.Println("Pretty print:")
	treeTree.PrettyPrint()

	var val string
	subtree, found := treeTree.Find(2)
	if found {
		val, found = subtree.Find("b")
	}
	fmt.Printf("Find \"s\" in subtree 2: %v (found: %t)\n", val, found)

}

/*

## How to run the code

This [code](https://github.com/appliedgo/generictree) runs with Go 1.21 or later. It also runs fine in the [Go Playground](https://go.dev/play/p/Jw9f9zM_bUi).


## Conclusion

Turning an existing container data type into a generic one has only few surprises. Hey, I told you it will be boring!

With a few checks in mind, you should be ready for generizing... generalizing... genericizing... genericking... uh, whatever... your existing container data types.

- Review all the operations your code applies to the original types. If these operations apply to a certain kind of data type only, your generic type needs a type constraint.
- Look through your `fmt.Printf` statements. Most likely, you will need to change a few type-specific placeholders to a general `%v` to avoid errors.
- Look for return statements that return a zero value. Typically, these occur when returning a non-nil error.\
  Example: `return "", errors.New(...)`. \
  Use one of the workaround shown above:
	- Workaround 1: declare a variable of type T, which defaults to the type's zero value. Return that variable.
	- Workaround 2: use `*new(T)`, which instantiates T, returns a pointer, and dereferences that pointer. The result is a zero value of T. Return that result.

(See the tree code above for working examples.)

In summary, I am pleased about how easy the conversion process turned out to be, and also how readable the result is. Once generics are included in an official release, workarounds [like the ones I described in another article](https://appliedgo.net/generics) are not required anymore.

That's it. Happy generic coding! ʕ◔ϖ◔ʔ

___

*Trees and background image courtesy of artists at Pixabay*

Changelog

2023-08-22

- Updated the code to work with Go 1.21. New: The `cmp` package. Obsolete: the `constraints` package. Link to the playground updated accordingly.
- Added missing link to the github repo of this article.

2022-01-04

- Updated the code from go2go version of May 2021 to the current dev branch (which is a pre-release version of Go 1.18). The code is now compatible with Go 1.18. The playground link now opens the current dev branch rather than the (obsolete) go2go Playground.
- I also took the chance to change "we" to "I" to match the title of the article.
*/
