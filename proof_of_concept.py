from __future__ import annotations
from dataclasses import dataclass
from typing import List, Iterator
from copy import copy, deepcopy
import random
from tabulate import tabulate


@dataclass
class Point:
    x: int
    y: int


@dataclass
class Rect:
    min: Point
    max: Point


class Board:
    def __init__(self, width: int, height: int, chars: List[List[str]] = None):
        self.width = width
        self.height = height

        if chars:
            self.chars = chars
        else:
            self.clear()

    def clear(self):
        self.chars = [[None for y in range(self.height)]
                      for x in range(self.width)]

    def clone(self) -> Board:
        return Board(self.width, self.height, chars=deepcopy(self.chars))

    # Scans an area for instances of char
    # and returns a list of their locations.
    def scan(self, char: str, area: Rect = None) -> Iterator[Point]:
        if not area:
            area = Rect(
                Point(0, 0),
                Point(self.width-1, self.height-1)
            )

        for x in range(area.min.x, area.max.x+1):
            for y in range(area.min.y, area.max.y+1):
                if self.chars[x][y] == char:
                    yield Point(x, y)

    def scan_adjacent(self, char: str, slot: Point) -> Iterator[Point]:
        area = Rect(
            Point(
                max(slot.x - 1, 0),
                max(slot.y - 1, 0)
            ),
            Point(
                min(slot.x + 1, self.width - 1),
                min(slot.y + 1, self.height - 1)
            )
        )

        for p in self.scan(char, area):
            if not (slot.x == p.x and slot.y == p.y):
                yield slot


class Filler:
    board: Board

    def __init__(self, board: Board):
        self.board = board

    def fill(self, words: List[str]) -> List[List[Point]]:
        paths = []
        for word in words:
            path = self.__insert_word(word)
            if not path:
                return None

            paths.append(path)

        return paths

    def __insert_word(self, word: str) -> List[Point]:
        # Find a random empty slot to insert the word at.
        empty_slots = list(self.board.scan(None))
        if not empty_slots:
            return False

        slot = random.choice(empty_slots)

        path = []
        for char in word:
            self.board.chars[slot.x][slot.y] = char
            path.append(slot)

            # Find all adjacent empty slots and choose
            # a random one to insert the next letter at.
            empty_slots = list(self.board.scan_adjacent(None, slot))
            if not empty_slots:
                return None

            slot = random.choice(empty_slots)

        return path


class Validator:
    def __init__(self, board: Board):
        self.board = board

    def validate(self, words: List[str], paths: List[List[Point]]) -> bool:
        # Find all possible starting locations of each word,
        # and check each location to determine whether
        # the word can be selected in any other path
        # than the one that's been defined for it.
        for i, word in enumerate(words):
            for slot in self.board.scan(word[0]):
                if self.__check_collision(word[1:], paths[i], slot):
                    return False

        return True

    def __check_collision(self, word: str, path: List[Point], pos: Point, out_of_path: bool = False) -> bool:
        if len(word) == 0:
            return out_of_path

        for slot in self.board.scan_adjacent(word[0], pos):
            if slot not in path:
                out_of_path = True

            if self.__check_collision(word[1:], path, slot, out_of_path):
                return True

        return False


charset = 'אבגדהוזחטיךכלםמןנסעףפץצקרשת'
board = Board(5, 5)
words = ["שלומ", "מתוק", "מה", "איתך"]

while True:
    board.clear()

    paths = Filler(board).fill(words)
    if not paths:
        continue

    if not Validator(board).validate(words, paths):
        continue

    # Add noise.
    while True:
        noisy_board = board.clone()
        for slot in noisy_board.scan(None):
            noisy_board.chars[slot.x][slot.y] = random.choice(charset)

        if Validator(noisy_board).validate(words, paths):
            break

    board = noisy_board

    break

table = [[board.chars[x][y] for x in range(board.width)]
         for y in range(board.height)]
print(tabulate(table, tablefmt="grid"))
