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

	//10 heartbeats per second
	heartbeatInverval = 100
	//time between 250 and 400 ms
	timeInterval = 150 + 1
	minTime      = 250
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
	currentTerm       int
	votedFor          int
	role              Role // role of server (0: follower, 1: candidate, 2: leader)
	votesReceived     int
	leaderId          int
	electionTimer     *time.Timer
	lastHeartbeatSent time.Time
	//heartbeatTimer *time.Timer
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	// Your code here (2A).
	rf.mu.Lock()
	var term int
	var isleader bool
	term = rf.currentTerm
	isleader = rf.role == Leader
	rf.mu.Unlock()

	return term, isleader
}

func (rf *Raft) persist() {

}

func (rf *Raft) readPersist(data []byte) {
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
	rf.electionTimer = time.NewTimer(time.Duration(minTime+rand.Intn(timeInterval)) * time.Millisecond)
	rf.votesReceived = 0
	rf.leaderId = -1

	// Your initialization code here (2A, 2B, 2C).
	go rf.routine() // start routine for server

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	return rf
}

func (rf *Raft) convertToFollower() {
	rf.role = Follower
	rf.electionTimer.Reset(time.Duration(minTime+rand.Intn(timeInterval)) * time.Millisecond)
	rf.votedFor = -1
}

func (rf *Raft) convertToLeader() {
	rf.role = Leader
	rf.electionTimer.Stop()
}

func (rf *Raft) convertToCandidate() {
	rf.role = Candidate
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	fmt.Println("[REQUESTVOTE] me:", rf.me, "currentTerm:", rf.currentTerm, "args:", args)
	rf.mu.Lock()
	defer rf.mu.Unlock()

	//Reply false if term < currentTerm (§5.1)
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		return
	}

	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.convertToFollower()
	}

	if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
		rf.votedFor = args.CandidateId
		reply.VoteGranted = true // true because term is greater than or equal to currentTerm
	}

	reply.Term = rf.currentTerm
	rf.electionTimer.Reset(time.Duration(minTime+rand.Intn(timeInterval)) * time.Millisecond)
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.Success = false // false because term is less than currentTerm
		return
	}

	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.convertToFollower()
	}

	rf.electionTimer.Reset(time.Duration(minTime+rand.Intn(timeInterval)) * time.Millisecond)
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) routine() {
	for {
		select {
		case <-rf.electionTimer.C:
			if rf.role == Follower {
				rf.convertToCandidate()
			}
			rf.startElection()
		default:
			if rf.role == Leader && (time.Since(rf.lastHeartbeatSent) > heartbeatInverval) {
				rf.lastHeartbeatSent = time.Now()
				rf.broadcastHeartbeat()
				time.Sleep(heartbeatInverval * time.Millisecond)
			}
		}
	}
}

func (rf *Raft) startElection() {
	rf.currentTerm++    //Increment currentTerm
	rf.votedFor = rf.me // vote for self
	rf.votesReceived = 1
	rf.electionTimer.Reset(time.Duration(minTime+rand.Intn(timeInterval)) * time.Millisecond) //Reset election timer
	go rf.broadcastRequestVote()                                                              // start broadcasting request vote RPC messages
}

func (rf *Raft) broadcastRequestVote() {
	for i := 0; i < len(rf.peers); i++ {
		if i != rf.me {
			rf.mu.Lock()
			requestVoteArgs := &RequestVoteArgs{
				Term:        rf.currentTerm,
				CandidateId: rf.me,
			}
			requestVoteReply := &RequestVoteReply{}
			rf.mu.Unlock()

			go func(i int) {
				ok := rf.sendRequestVote(i, requestVoteArgs, requestVoteReply)
				if ok {
					rf.mu.Lock()
					if requestVoteReply.VoteGranted {
						rf.votesReceived++
						if rf.votesReceived > len(rf.peers)/2 {
							fmt.Println("[LEADER ELECTED]", rf.me)
							rf.convertToLeader()
						}

					} else if requestVoteReply.Term > rf.currentTerm {
						fmt.Println("requestVoteReply.Term > rf.currentTerm")
						rf.convertToFollower()
					}
					rf.mu.Unlock()
				}
			}(i)
		}
	}
}

func (rf *Raft) broadcastHeartbeat() {
	for i := 0; i < len(rf.peers); i++ {
		if i != rf.me {
			go func(i int) {
				rf.mu.Lock()
				appendEntriesArgs := &AppendEntriesArgs{
					Term:     rf.currentTerm,
					LeaderId: rf.me,
				}
				rf.mu.Unlock()
				appendEntriesReply := &AppendEntriesReply{}
				ok := rf.sendAppendEntries(i, appendEntriesArgs, appendEntriesReply)

				if ok && !appendEntriesReply.Success && appendEntriesReply.Term > rf.currentTerm {
					rf.mu.Lock()
					rf.currentTerm = appendEntriesReply.Term
					rf.convertToFollower()

					rf.mu.Unlock()
				}
			}(i)
		}
	}
}

func (rf *Raft) listenHeartBeat() {
	for {
		if rf.role != Follower {
			break
		}

	}
}
