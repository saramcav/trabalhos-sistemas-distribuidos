package raft

import (
	"fmt"
	"labrpc"
	"math/rand"
	"sync"
	"time"
)

type Role int

const (
	Follower Role = iota
	Candidate
	Leader

	heartBeatInterval = 200
	timeoutNumFollower = 1000
	timeoutNumElection = 300
)

type AppendEntriesArgs struct {
	Term     int
	LeaderId int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	currentTerm   int
	votedFor      int
	role          Role  // role of server (0: follower, 1: candidate, 2: leader)
	timer         int64 // timer for election timeout
	startTime     int64 // start time of timer
	votesReceived int
	leaderId      int
	isDead        bool
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isleader bool
	// Your code here (2A).
	term = rf.currentTerm
	isleader = rf.role == Leader

	return term, isleader
}

func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := gob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

func (rf *Raft) readPersist(data []byte) {
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := gob.NewDecoder(r)
	// d.Decode(&rf.xxx)
	// d.Decode(&rf.yyy)
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
}

type RequestVoteArgs struct {
	Term        int
	CandidateId int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).

	return index, term, isLeader
}

func (rf *Raft) Kill() {
	fmt.Println("Killing server ", rf.me)
	rf.isDead = true // set isDead to true
}

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.role = Follower
	rf.timer = randomTimeFollower()
	rf.startTime = currentTime()
	rf.votesReceived = 0
	rf.leaderId = -1
	rf.isDead = false

	// Your initialization code here (2A, 2B, 2C).
	go rf.routine() // start routine for server

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	return rf
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	//fmt.Println("[RequestVote] me:", rf.me, "currentTerm:", rf.currentTerm, "args:", args)
	rf.mu.Lock()
	defer rf.mu.Unlock()

	//Reply false if term < currentTerm (§5.1)
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.VoteGranted = false

	} else if rf.votedFor == -1 || rf.votedFor == args.CandidateId{
		rf.votedFor = args.CandidateId
		rf.currentTerm = args.Term
		rf.role = Follower
		rf.startTime = currentTime()
		rf.timer = randomTimeFollower()

		reply.Term = args.Term
		reply.VoteGranted = true // true because term is greater than or equal to currentTerm
	}
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.Success = false // false because term is less than currentTerm
		return
	}

	if rf.role != Follower && !rf.isDead {
		if args.LeaderId != rf.leaderId {
			rf.leaderId = args.LeaderId
		}
		rf.role = Follower
		rf.timer = randomTimeFollower()
	}

	rf.currentTerm = args.Term
	reply.Term = args.Term
	reply.Success = true // true because term is greater than or equal to currentTerm

	rf.startTime = currentTime() // reset start time of timer
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) routine() {
	for {
		if rf.isDead {
			return
		}

		if rf.role == Follower {
			if ((currentTime() - rf.startTime) > rf.timer) && rf.votedFor == -1{
				fmt.Println("Server ", rf.me, " timed out and became candidate.")
				rf.mu.Lock()
				rf.role = Candidate
				rf.mu.Unlock()
			}
		}
		if rf.role == Candidate {
			go rf.startElection()
		}
		if rf.role == Leader {
			go rf.broadcastHeartbeat()
		}

	}
}

func (rf *Raft) startElection() {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	rf.currentTerm++
	rf.votedFor = rf.me // vote for self
	rf.votesReceived = 1
	rf.startTime = currentTime() // reset start time of timer
	rf.timer = randomTimeElection()
	go rf.broadcastRequestVote() // start broadcasting request vote messages
}

func (rf *Raft) broadcastRequestVote() {
	for i := 0; i < len(rf.peers); i++ {
		if i != rf.me {
			requestVoteArgs := &RequestVoteArgs{
				Term:        rf.currentTerm,
				CandidateId: rf.me,
			}
			requestVoteReply := &RequestVoteReply{}

			go func(i int) {
				ok := rf.sendRequestVote(i, requestVoteArgs, requestVoteReply)
				if ok {
					rf.mu.Lock()
					defer rf.mu.Unlock()

					if rf.role != Candidate {
						return
				   	}

					if (currentTime() - rf.startTime) > rf.timer {
						return
					}

					if requestVoteReply.VoteGranted {
						rf.votesReceived++
						if rf.votesReceived > len(rf.peers)/2 {
							fmt.Println("New leader: ", rf.me)
							rf.role = Leader
							rf.votedFor = -1
							rf.votesReceived = 0
							rf.leaderId = rf.me
						}

					} else if requestVoteReply.Term > rf.currentTerm {
						fmt.Println("requestVoteReply.Term > rf.currentTerm")
						rf.currentTerm = requestVoteReply.Term
						rf.role = Follower
						rf.votedFor = -1
						rf.votesReceived = 0
						//adicionei
						rf.startTime = currentTime()
						rf.timer = randomTimeFollower()
					}
				}
			}(i)
		}
	}
}

func (rf *Raft) broadcastHeartbeat() {
	for !rf.isDead {
		if rf.role != Leader {
			return
		}

		appendEntriesArgs := &AppendEntriesArgs{
			Term:     rf.currentTerm,
			LeaderId: rf.me,
		}

		for i := 0; i < len(rf.peers); i++ {
			if i != rf.me {
				go func(i int) {
					appendEntriesReply := &AppendEntriesReply{}
					ok := rf.sendAppendEntries(i, appendEntriesArgs, appendEntriesReply)

					rf.mu.Lock()
					defer rf.mu.Unlock()

					if ok && !appendEntriesReply.Success && appendEntriesReply.Term > rf.currentTerm {
						rf.mu.Lock()
						rf.currentTerm = appendEntriesReply.Term
						rf.role = Follower
						rf.votedFor = -1
						rf.timer = randomTimeFollower()
						rf.startTime = currentTime() // reset start time of timer
						rf.mu.Unlock()
					}
				}(i)
			}
		}
		time.Sleep(heartBeatInterval * time.Millisecond)
	}
}

func (rf *Raft) listenHeartBeat() {
	for {
		if rf.role != Follower || rf.isDead {
			break
		}

	}
}

func randomTimeFollower() int64 {
	return int64(timeoutNumFollower + rand.Intn(timeoutNumFollower+1))
}

func randomTimeElection() int64 {	
	return int64(timeoutNumElection + rand.Intn(timeoutNumElection+1))
}

func currentTime() int64 {
	return int64(time.Now().UnixNano() / 1000000)
}
