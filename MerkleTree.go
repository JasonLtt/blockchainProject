package main

import "crypto/sha256"

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	//如果数据的长度为奇数，多添加一次末尾数据
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	//构造叶子节点
	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	//构造Merkle树
	for i := 0; i < len(data)/2; i++ {
		var NewLevel []MerkleNode
		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			NewLevel = append(NewLevel, *node)
		}
		nodes = NewLevel
	}

	//取出Merkle根的hash值
	mTree := MerkleTree{&nodes[0]}

	return &mTree
}

//构造Merkle树，并加以hash
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	mNode := MerkleNode{}

	if left == nil && right == nil {//叶子结点
		hash := sha256.Sum256(data)
		mNode.Data = hash[:]
	} else {//非叶子节点
		preveHash := append(left.Data, right.Data...)
		hash := sha256.Sum256(preveHash)
		mNode.Data = hash[:]
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode
}
